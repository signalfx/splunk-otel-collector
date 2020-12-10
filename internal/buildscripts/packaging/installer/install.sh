#!/bin/sh
# Copyright 2020 Splunk, Inc.
# Copyright The OpenTelemetry Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# A convenience script to install the collector package on any of our supported
# distros.  NOT recommended for production use.

set -euf

get_distro() {
  local distro="$(. /etc/os-release 2>/dev/null && echo ${ID:-} || true)"

  echo "$distro"
}

get_distro_version() {
  local version="$(. /etc/os-release 2>/dev/null && echo ${VERSION_ID:-} || true)"

  echo "$version"
}

get_distro_codename() {
  local distro="$( get_distro )"
  local codename="$(. /etc/os-release 2>/dev/null && echo ${VERSION_CODENAME:-} || true)"

  if [ "$distro" = "debian" ] && [ -z "$codename" ]; then
    case "$( get_distro_version )" in
      8)
        codename="jessie"
        ;;
      9)
        codename="stretch"
        ;;
      10)
        codename="buster"
        ;;
      *)
        codename=""
        ;;
    esac
  fi

  echo "$codename"
}

distro="$( get_distro )"
distro_version="$( get_distro_version )"
distro_codename="$( get_distro_codename )"
repo_base="https://splunk.jfrog.io/splunk"
deb_repo_base="${repo_base}/otel-collector-deb"
rpm_repo_base="${repo_base}/otel-collector-rpm"
debian_gpg_key_url="${deb_repo_base}/splunk-B3CD4420.gpg"
yum_gpg_key_url="${rpm_repo_base}/splunk-B3CD4420.pub"
td_agent_repo_base="https://packages.treasuredata.com"
td_agent_gpg_key_url="${td_agent_repo_base}/GPG-KEY-td-agent"
splunk_config_path="/etc/otel/collector/splunk_config.yaml"
splunk_env_path="/etc/otel/collector/splunk_env"

default_stage="release"
default_realm="us0"
default_ballast_size="683"
default_collector_version="latest"
default_td_agent_version="4.0.1"
default_td_agent_version_jessie="3.3.0-1"
default_td_agent_version_stretch="3.7.1-0"
default_service_user="splunk-otel-collector"
default_service_group="splunk-otel-collector"

if [ "$distro_codename" = "stretch" ]; then
  default_td_agent_version="$default_td_agent_version_stretch"
elif [ "$distro_codename" = "jessie" ]; then
  default_td_agent_version="$default_td_agent_version_jessie"
fi

repo_for_stage() {
  local repo_url=$1
  local stage=$2
  echo "$repo_url/$stage"
}

download_file_to_stdout() {
  local url=$1

  if command -v curl > /dev/null; then
    curl -sSL $url
  elif command -v wget > /dev/null; then
    wget -O - -o /dev/null $url
  else
    echo "Either curl or wget must be installed to download $url" >&2
    exit 1
  fi
}

request_access_token() {
  local access_token=
  while [ -z "$access_token" ]; do
    read -p "Please enter your Splunk access token: " access_token
  done
  echo "$access_token"
}

verify_access_token() {
  local access_token="$1"
  local ingest_url="$2"
  local insecure="$3"

  if command -v curl > /dev/null; then
    api_output=$(curl \
      -d '[]' \
      -H "X-Sf-Token: $access_token" \
      -H "Content-Type:application/json" \
      -X POST \
      $([ "$insecure" = "true" ] && echo -n "--insecure") \
      "$ingest_url"/v2/event 2>/dev/null)
  elif command -v wget > /dev/null; then
    api_output=$(wget \
      --header="Content-Type: application/json" \
      --header="X-Sf-Token: $access_token" \
      --post-data='[]' \
      $([ "$insecure" = "true" ] && echo -n "--no-check-certificate") \
      -O - \
      -o /dev/null \
      "$ingest_url"/v2/event)
    if [ $? -eq 5 ]; then
      echo "TLS cert for Splunk ingest could not be verified, does your system have TLS certs installed?" >&2
      exit 1
    fi
  else
    echo "Either curl or wget is required to verify the access token" >&2
    exit 1
  fi

  if [ "$api_output" = "\"OK\"" ]; then
    true
  else
    echo "$api_output"
    false
  fi
}

download_debian_key() {
  local key_url="$1"
  local key_path="$2"

  if ! download_file_to_stdout "$key_url" > $key_path; then
    echo "Could not get Debian GPG signing key from $key_url" >&2
    exit 1
  fi
  chmod 644 $key_path
}

install_collector_apt_repo() {
  local stage="$1"
  local trusted_flag=
  if [ "$stage" = "test" ]; then
    trusted_flag="[trusted=yes]"
  fi

  download_debian_key "$debian_gpg_key_url" "/etc/apt/trusted.gpg.d/splunk.gpg"
  echo "deb $trusted_flag $deb_repo_base $stage main" > /etc/apt/sources.list.d/splunk-otel-collector.list
}

install_td_agent_apt_repo() {
  local td_agent_version="$1"
  local td_agent_major_version="$( echo $td_agent_version | cut -d '.' -f1 )"

  if ! download_file_to_stdout "$td_agent_gpg_key_url" | apt-key add -; then
    echo "Could not download Debian GPG key from $td_agent_gpg_key_url" >&2
    exit 1
  fi

  echo "deb ${td_agent_repo_base}/${td_agent_major_version}/${distro}/${distro_codename} $distro_codename contrib" > /etc/apt/sources.list.d/td_agent.list
}

install_apt_package() {
  local package_name="$1"
  local version="$2"

  if [ "$version" = "latest" ]; then
    version=""
  elif [ -n "$version" ]; then
    version="=${version}"
  fi

  apt-get -y update
  apt-get -y install ${package_name}${version}
}

install_collector_yum_repo() {
  local stage="$1"
  local repo_path="${2:-/etc/yum.repos.d}"
  local gpgcheck="1"

  if [ "$stage" = "test" ]; then
    gpgcheck="0"
  fi

  cat <<EOH > ${repo_path}/splunk-otel-collector.repo
[splunk-otel-collector]
name=Splunk OpenTelemetry Collector Repository
baseurl=$( repo_for_stage $rpm_repo_base $stage )/\$basearch
gpgcheck=$gpgcheck
repo_gpgcheck=$gpgcheck
gpgkey=$yum_gpg_key_url
enabled=1
EOH
}

install_td_agent_yum_repo() {
  local td_agent_version="$1"
  local repo_path="${2:-/etc/yum.repos.d}"
  local td_agent_major_version="$( echo $td_agent_version | cut -d '.' -f1 )"
  local releasever="$( echo "$distro_version" | cut -d '.' -f1 )"

  if [ "$distro" = "amzn" ]; then
    distro="amazon"
  else
    distro="redhat"
  fi

  cat <<EOH > ${repo_path}/td_agent.repo
[td_agent]
name=TreasureData Repository
baseurl=${td_agent_repo_base}/${td_agent_major_version}/${distro}/${releasever}/\$basearch
gpgcheck=1
gpgkey=$td_agent_gpg_key_url
enabled=1
EOH
}

install_yum_package() {
  local package_name="$1"
  local version="$2"

  if [ "$version" = "latest" ]; then
    version=""
  elif [ -n "$version" ]; then
    version="-${version}"
  fi

  yum install -y ${package_name}${version}
}

ensure_not_installed() {
  if [ -e /etc/otel/collector ]; then
    echo "The collector directory /etc/otel/collector already exists which implies that the collector has already been installed.  Please remove this directory to proceed." >&2
    exit 1
  fi
}

configure_access_token() {
  local access_token=$1
  local service_user=$2
  local service_group=$3

  mkdir -p /etc/otel/collector
  printf "SPLUNK_ACCESS_TOKEN=%s\n" "$access_token" >> $splunk_env_path
  chmod 600 $splunk_env_path
  chown $service_user:$service_group $splunk_env_path
}

configure_realm() {
  local realm=$1
  local service_user=$2
  local service_group=$3

  mkdir -p /etc/otel/collector
  printf "SPLUNK_REALM=%s\n" "$realm" >> $splunk_env_path
  chmod 600 $splunk_env_path
  chown $service_user:$service_group $splunk_env_path
}

configure_ballast() {
  local ballast=$1
  local service_user=$2
  local service_group=$3

  mkdir -p /etc/otel/collector
  printf "SPLUNK_BALLAST_SIZE_MIB=%s\n" "$ballast" >> $splunk_env_path
  chmod 600 $splunk_env_path
  chown $service_user:$service_group $splunk_env_path
}

create_user_group() {
  local user="$1"
  local group="$2"

  getent group $group >/dev/null 2>&1 || \
    groupadd --system $group

  getent passwd $user >/dev/null 2>&1 || \
    useradd --system --no-user-group --home-dir /etc/otel/collector --no-create-home --shell $(command -v nologin) --groups $group $user
}

configure_service_owner() {
  local service_user="$1"
  local service_group="$2"
  local tmpfile_path="/etc/tmpfiles.d/splunk-otel-collector.conf"
  local override_path="/etc/systemd/system/splunk-otel-collector.service.d/service-owner.conf"

  systemctl stop splunk-otel-collector

  mkdir -p $(dirname $tmpfile_path)
  cat <<EOH > $tmpfile_path
D /run/splunk-otel-collector 0755 ${service_user} ${service_group} - -
EOH
  systemd-tmpfiles --create --remove $tmpfile_path

  mkdir -p $(dirname $override_path)
  cat <<EOH > $override_path
[Service]
User=${service_user}
Group=${service_group}
EOH

  systemctl daemon-reload
}

install() {
  local stage="$1"
  local realm="$2"
  local access_token="$3"
  local insecure="$4"
  local collector_version="$5"
  local td_agent_version="$6"
  local ingest_url="https://ingest.${realm}.signalfx.com"

  if [ -z $access_token ]; then
    access_token=$(request_access_token)
  fi

  if [ "${NO_SPLUNK_TOKEN_VERIFY:-}" != "yes" ] && ! verify_access_token "$access_token" "$ingest_url" "$insecure"; then
    echo "Your access token could not be verified. This may be due to a network connectivity issue." >&2
    exit 1
  fi

  case "$distro" in
    ubuntu|debian)
      if [ -z "$distro_codename" ]; then
        echo "The distribution codename could not be determined" >&2
        exit 1
      fi
      apt-get -y update
      apt-get -y install apt-transport-https gnupg
      install_collector_apt_repo "$stage"
      install_apt_package "splunk-otel-collector" "$collector_version"
      install_td_agent_apt_repo "$td_agent_version"
      if [ "$(echo "$td_agent_version" | cut -d'-' -f2)" = "$td_agent_version" ]; then
        td_agent_version="$td_agent_version-1"
      fi
      install_apt_package "td-agent" "$td_agent_version"
      ;;
    amzn|centos|ol|rhel)
      if [ -z "$distro_version" ]; then
        echo "The distribution version could not be determined" >&2
        exit 1
      fi
      install_collector_yum_repo "$stage"
      install_yum_package "splunk-otel-collector" "$collector_version"
      install_td_agent_yum_repo "$td_agent_version"
      install_yum_package "td-agent" "$td_agent_version"
      ;;
    *)
      echo "Your distro ($distro) is not supported or could not be determined" >&2
      exit 1
      ;;
  esac
}

usage() {
  cat <<EOH >&2
Usage: $0 [options] [access_token]

Installs the Splunk OpenTelemetry Collector from the package repos.
If access_token is not provided, it will prompted for on stdin.

Options:

  --collector-version <version>     The splunk-otel-collector package version to install (default: "latest")
  --realm <us0|us1|eu0|...>         The Splunk realm to use (default: "$default_realm")
  --ballast <ballast size>          How much memory in MIB to allocate to the ballast (default: "$default_ballast_size")
                                    This should be set to 1/3 to 1/2 of configured memory
  --service-user <user>             Set the user for the splunk-otel-collector service (default: "splunk-otel-collector")
                                    The user will be created if it does not exist
  --service-group <group>           Set the group for the splunk-otel-collector service (default: "splunk-otel-collector")
                                    The group will be created if it does not exist
  --td-agent-version <version>      The td-agent (fluentd) package version to install (default: "$default_td_agent_version")
  --test                            Use the test package repo instead of the primary
  --beta                            Use the beta package repo instead of the primary
  --                                Use -- if your access_token starts with -

EOH
  exit 0
}

parse_args_and_install() {
  local stage="$default_stage"
  local realm="$default_realm"
  local ballast="$default_ballast_size"
  local access_token=
  local insecure=
  local collector_version="$default_collector_version"
  local service_user="$default_service_user"
  local service_group="$default_service_group"
  local td_agent_version="$default_td_agent_version"

  while [ -n "${1-}" ]; do
    case $1 in
      --beta)
        stage="beta"
        ;;
      --test)
        stage="test"
        ;;
      --api-url)
        api_url="$2"
        shift 1
        ;;
      --realm)
        realm="$2"
        shift 1
        ;;
      --ballast)
        ballast="$2"
        shift 1
        ;;
      --insecure)
        insecure="true"
        ;;
      --collector-version)
        collector_version="$2"
        shift 1
        ;;
      --service-user)
        service_user="$2"
        shift 1
        ;;
      --service-group)
        service_group="$2"
        shift 1
        ;;
      --td-agent-version)
        td_agent_version="$2"
        shift 1
        ;;
      --)
        access_token="$2"
        shift 1
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      -*)
        echo "Unknown option $1" >&2
        usage
        exit 1
        ;;
      *)
        if [ -z "$access_token" ]; then
          access_token=$1
        else
          echo "Unknown argument $1" >&2
          usage
          exit 1
        fi
        ;;
    esac
    shift 1
  done

  ensure_not_installed

  echo "Splunk OpenTelemetry Collector Version: ${collector_version}"
  echo "Realm: $realm"
  echo "Ballast Size in MIB: $ballast"
  echo "TD Agent Fluentd Version: $td_agent_version"

  install "$stage" "$realm" "$access_token" "$insecure" "$collector_version" "$td_agent_version"

  create_user_group "$service_user" "$service_group"
  configure_service_owner "$service_user" "$service_group"

  configure_access_token "$access_token" "$service_user" "$service_group"
  configure_realm "$realm" "$service_user" "$service_group"
  configure_ballast "$ballast" "$service_user" "$service_group"

  # disable fluentd from starting with default config
  systemctl stop td-agent
  if [ -f /etc/td-agent/td-agent.conf ]; then
    mv -f /etc/td-agent/td-agent.conf /etc/td-agent/td-agent.conf.default
  fi
  systemctl disable td-agent

  systemctl start splunk-otel-collector

  cat <<EOH
The Splunk OpenTelemetry Collector has been successfully installed.

Make sure that your system's time is relatively accurate or else datapoints may not be accepted.

The collector's main configuration file is located at $splunk_config_path,
and the environment file is located at $splunk_env_path.

If either $splunk_config_path or $splunk_env_path are modified, the collector service
must be restarted to apply the changes by running the following command as root:

  systemctl restart splunk-otel-collector

EOH
  exit 0
}

parse_args_and_install $@
