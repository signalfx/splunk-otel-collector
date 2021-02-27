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

collector_config_dir="/etc/otel/collector"
collector_config_path="${collector_config_dir}/splunk_config_linux.yaml"
collector_env_path="${collector_config_dir}/splunk_env"
distro="$( get_distro )"
distro_codename="$( get_distro_codename )"
distro_version="$( get_distro_version )"
repo_base="https://splunk.jfrog.io/splunk"

deb_repo_base="${repo_base}/otel-collector-deb"
debian_gpg_key_url="${deb_repo_base}/splunk-B3CD4420.gpg"

rpm_repo_base="${repo_base}/otel-collector-rpm"
yum_gpg_key_url="${rpm_repo_base}/splunk-B3CD4420.pub"

fluent_config_dir="${collector_config_dir}/fluentd"
fluent_config_path="${fluent_config_dir}/fluent.conf"
td_agent_repo_base="https://packages.treasuredata.com"
td_agent_gpg_key_url="${td_agent_repo_base}/GPG-KEY-td-agent"

default_stage="release"
default_realm="us0"
default_memory_size="512"

default_collector_version="latest"
default_td_agent_version="4.1.0"
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

  apt-get -y install -o Dpkg::Options::="--force-confold" ${package_name}${version}
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

  if command -v yum >/dev/null 2>&1; then
    yum install -y ${package_name}${version}
  else
    dnf install -y ${package_name}${version}
  fi
}

ensure_not_installed() {
  for agent in otelcol td-agent; do
    if command -v $agent >/dev/null 2>&1; then
        echo "An agent binary already exists at $( command -v $agent ) which implies that the agent has already been installed." >&2
        echo "Please uninstall the agent and re-run this script." >&2
      exit 1
    fi
  done
}

configure_env_file() {
  local key="$1"
  local value="$2"
  local env_file="$3"

  mkdir -p "$(dirname $env_file)"

  if [ -f "$env_file" ] && grep -q "^${key}=" "$env_file"; then
    sed -i "s|^${key}=.*|${key}=${value}|" "$env_file"
  else
    echo "${key}=${value}" >> "$env_file"
  fi
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

  chown root:root $override_path
  chmod 644 $override_path
  systemctl daemon-reload
}

configure_fluentd_service() {
  local override_src_path="$fluent_config_dir/splunk-otel-collector.conf"
  local override_dest_path="/etc/systemd/system/td-agent.service.d/splunk-otel-collector.conf"

  if [ -f "$override_src_path" ]; then
    systemctl stop td-agent
    mkdir -p $(dirname $override_dest_path)
    cp -f $override_src_path $override_dest_path
    chown root:root $override_dest_path
    chmod 644 $override_dest_path
    systemctl daemon-reload
  fi
}

fluent_plugin_installed() {
  local name="1"

  td-agent-gem list "$name" --exact | grep -q "$name"
}

install_fluent_plugin() {
  local name="$1"
  local version="${2:-}"

  if [ -n "$version" ]; then
    td-agent install "$name" --version "$version"
  else
    td-agent install "$name"
  fi
}

install() {
  local stage="$1"
  local collector_version="$2"
  local td_agent_version="$3"

  case "$distro" in
    ubuntu|debian)
      if [ -z "$distro_codename" ]; then
        echo "The distribution codename could not be determined" >&2
        exit 1
      fi
      apt-get -y update
      apt-get -y install apt-transport-https gnupg
      install_collector_apt_repo "$stage"
      apt-get -y update
      install_apt_package "splunk-otel-collector" "$collector_version"
      if [ -n "$td_agent_version" ]; then
        if [ "$distro_codename" != "stretch" ] && [ "$distro_codename" != "jessie" ]; then
          td_agent_version="${td_agent_version}-1"
        fi
        install_td_agent_apt_repo "$td_agent_version"
        apt-get -y update
        install_apt_package "td-agent" "$td_agent_version"
        systemctl stop td-agent
      fi
      ;;
    amzn|centos|ol|rhel)
      if [ -z "$distro_version" ]; then
        echo "The distribution version could not be determined" >&2
        exit 1
      fi
      install_collector_yum_repo "$stage"
      install_yum_package "splunk-otel-collector" "$collector_version"
      if [ -n "$td_agent_version" ]; then
        install_td_agent_yum_repo "$td_agent_version"
        install_yum_package "td-agent" "$td_agent_version"
        systemctl stop td-agent
      fi
      ;;
    *)
      echo "Your distro ($distro) is not supported or could not be determined" >&2
      exit 1
      ;;
  esac
}

uninstall() {
  case "$distro" in
    ubuntu|debian)
      for agent in splunk-otel-collector td-agent; do
        if command -v $agent >/dev/null 2>&1; then
          apt-get remove $agent 2>&1
          echo "Successfully removed $agent"
        else
          echo "Unable to locate $agent"
        fi
      done
      ;;
    amzn|centos|ol|rhel)
      for agent in splunk-otel-collector td-agent; do
        if command -v $agent >/dev/null 2>&1; then
          if command -v yum >/dev/null 2>&1; then
            yum remove $agent 2>&1
          else
            dnf remove $agent 2>&1
          fi
          echo "Successfully removed $agent"
        else
          echo "Unable to locate $agent"
        fi
      done
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

Installs the Splunk OpenTelemetry Connector for Linux from the package repos.
If access_token is not provided, it will be prompted for on stdin.

Options:

  --api-url <url>                   Set the api endpoint URL explicitly instead of the endpoint inferred from the specified realm
                                    (default: https://api.REALM.signalfx.com)
  --ballast <ballast size>          Set the ballast size explicitly instead of the value calculated from the --memory option
                                    This should be set to 1/3 to 1/2 of configured memory
  --beta                            Use the beta package repo instead of the primary
  --collector-version <version>     The splunk-otel-collector package version to install (default: "$default_collector_version")
  --hec-token <token>               Set the HEC token if different than the specified Splunk access_token
  --hec-url <url>                   Set the HEC endpoint URL explicitly instead of the endpoint inferred from the specified realm
                                    (default: https://ingest.REALM.signalfx.com/v1/log)
  --ingest-url <url>                Set the ingest endpoint URL explicitly instead of the endpoint inferred from the specified realm
                                    (default: https://ingest.REALM.signalfx.com)
  --memory <memory size>            Total memory in MIB to allocate to the collector; automatically calculates the ballast size
                                    (default: "$default_memory_size")
  --realm <us0|us1|eu0|...>         The Splunk realm to use (default: "$default_realm")
                                    The ingest, api, trace, and HEC endpoint URLs will automatically be inferred by this value
  --service-group <group>           Set the group for the splunk-otel-collector service (default: "$default_service_group")
                                    The group will be created if it does not exist
  --service-user <user>             Set the user for the splunk-otel-collector service (default: "$default_service_user")
                                    The user will be created if it does not exist
  --test                            Use the test package repo instead of the primary
  --trace-url <url>                 Set the trace endpoint URL explicitly instead of the endpoint inferred from the specified realm
                                    (default: https://ingest.REALM.signalfx.com/v2/trace)
  --uninstall                       Removes the Splunk OpenTelemetry Connector for Linux
  --with[out]-fluentd               Whether to install and configure fluentd to forward log events to the collector
                                    (default: --with-fluentd)
  --                                Use -- if access_token starts with -

EOH
  exit 0
}

parse_args_and_install() {
  local access_token=
  local api_url=
  local ballast=
  local collector_version="$default_collector_version"
  local hec_token=
  local hec_url=
  local ingest_url=
  local insecure=
  local memory="$default_memory_size"
  local realm="$default_realm"
  local service_group="$default_service_group"
  local stage="$default_stage"
  local service_user="$default_service_user"
  local td_agent_version="$default_td_agent_version"
  local trace_url=
  local uninstall="false"
  local with_fluentd="true"

  while [ -n "${1-}" ]; do
    case $1 in
      --api-url)
        api_url="$2"
        shift 1
        ;;
      --ballast)
        ballast="$2"
        shift 1
        ;;
      --beta)
        stage="beta"
        ;;
      --collector-version)
        collector_version="$2"
        shift 1
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      --hec-token)
        hec_token="$2"
        shift 1
        ;;
      --hec-url)
        hec_url="$2"
        shift 1
        ;;
      --ingest-url)
        ingest_url="$2"
        shift 1
        ;;
      --insecure)
        insecure="true"
        ;;
      --memory)
        memory="$2"
        shift 1
        ;;
      --realm)
        realm="$2"
        shift 1
        ;;
      --service-group)
        service_group="$2"
        shift 1
        ;;
      --service-user)
        service_user="$2"
        shift 1
        ;;
      --test)
        stage="test"
        ;;
      --trace-url)
        trace_url="$2"
        shift 1
        ;;
      --uninstall)
        uninstall="true"
        ;;
      --with-fluentd)
        with_fluentd="true"
        ;;
      --without-fluentd)
        with_fluentd="false"
        ;;
      --)
        access_token="$2"
        shift 1
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

  if [ "$uninstall" = true ]; then
      uninstall
      exit 0
  fi

  ensure_not_installed

  if [ -z "$access_token" ]; then
    access_token=$(request_access_token)
  fi

  if [ -z "$api_url" ]; then
    api_url="https://api.${realm}.signalfx.com"
  fi

  if [ -z "$ingest_url" ]; then
    ingest_url="https://ingest.${realm}.signalfx.com"
  fi

  if [ -z "$hec_token" ]; then
    hec_token="$access_token"
  fi

  if [ -z "$hec_url" ]; then
    hec_url="${ingest_url}/v1/log"
  fi

  if [ "$with_fluentd" != "true" ]; then
    td_agent_version=""
  fi

  if [ -z "$trace_url" ]; then
    trace_url="${ingest_url}/v2/trace"
  fi

  echo "Splunk OpenTelemetry Collector Version: ${collector_version}"
  if [ -n "$ballast" ]; then
    echo "Ballast Size in MIB: $ballast"
  else
    echo "Memory Size in MIB: $memory"
  fi
  echo "Realm: $realm"
  echo "Ingest Endpoint: $ingest_url"
  echo "API Endpoint: $api_url"
  echo "Trace Endpoint: $trace_url"
  echo "HEC Endpoint: $hec_url"
  if [ "$with_fluentd" = "true" ]; then
    echo "TD Agent (Fluentd) Version: $td_agent_version"
  fi

  if [ "${VERIFY_ACCESS_TOKEN:-true}" = "true" ] && ! verify_access_token "$access_token" "$ingest_url" "$insecure"; then
    echo "Your access token could not be verified. This may be due to a network connectivity issue." >&2
    exit 1
  fi

  install "$stage" "$collector_version" "$td_agent_version"

  create_user_group "$service_user" "$service_group"
  configure_service_owner "$service_user" "$service_group"

  configure_env_file "SPLUNK_ACCESS_TOKEN" "$access_token" "$collector_env_path"
  configure_env_file "SPLUNK_REALM" "$realm" "$collector_env_path"
  configure_env_file "SPLUNK_API_URL" "$api_url" "$collector_env_path"
  configure_env_file "SPLUNK_INGEST_URL" "$ingest_url" "$collector_env_path"
  configure_env_file "SPLUNK_TRACE_URL" "$trace_url" "$collector_env_path"
  configure_env_file "SPLUNK_HEC_URL" "$hec_url" "$collector_env_path"
  configure_env_file "SPLUNK_HEC_TOKEN" "$hec_token" "$collector_env_path"
  if [ -n "$ballast" ]; then
    configure_env_file "SPLUNK_BALLAST_SIZE_MIB" "$ballast" "$collector_env_path"
  else
    configure_env_file "SPLUNK_MEMORY_TOTAL_MIB" "$memory" "$collector_env_path"
  fi

  # ensure the collector service owner has access to the config dir
  chown -R $service_user:$service_group "$collector_config_dir"

  # ensure only the collector service owner has access to the env file
  chmod 600 "$collector_env_path"

  systemctl daemon-reload
  systemctl restart splunk-otel-collector

  if [ "$with_fluentd" = "true" ]; then
    # only start fluentd with our custom config to avoid port conflicts within the default config
    systemctl stop td-agent
    if [ -f "$fluent_config_path" ]; then
      configure_fluentd_service
      systemctl restart td-agent
    else
      if [ -f /etc/td-agent/td-agent.conf ]; then
        mv -f /etc/td-agent/td-agent.conf /etc/td-agent/td-agent.conf.bak
      fi
      systemctl disable td-agent
    fi
  fi

  cat <<EOH
The Splunk OpenTelemetry Connector for Linux has been successfully installed.

Make sure that your system's time is relatively accurate or else datapoints may not be accepted.

The collector's main configuration file is located at $collector_config_path,
and the environment file is located at $collector_env_path.

If either $collector_config_path or $collector_env_path are modified, the collector service
must be restarted to apply the changes by running the following command as root:

  systemctl restart splunk-otel-collector

EOH

  if [ "$with_fluentd" = "true" ] && [ -f "$fluent_config_path" ]; then
    cat <<EOH
Fluentd has been installed and configured to forward log events to the Splunk OpenTelemetry Collector.
By default, all log events with the @SPLUNK label will be forwarded to the collector.

The main fluentd configuration file is located at $fluent_config_path.
Custom input sources and configurations can be added to the ${fluent_config_dir}/conf.d/ directory.
All files with the .conf extension in this directory will automatically be included by fluentd.

Note: The fluentd service runs as the "td-agent" user.  When adding new input sources or configuration
files to the ${fluent_config_dir}/conf.d/ directory, ensure that the "td-agent" user has permissions
to access the new config files and the paths defined within.

If the fluentd configuration is modified or new config files are added, the fluentd service must be
restarted to apply the changes by running the following command as root:

  systemctl restart td-agent

EOH
  fi

  exit 0
}

parse_args_and_install $@
