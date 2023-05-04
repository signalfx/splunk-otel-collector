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
# distros.  Please refer to
# https://github.com/signalfx/splunk-otel-collector/tree/main/deployments for
# other installation methods (ansible, chef, puppet, etc.) that are
# more suitable for production environments.

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
      9)
        codename="stretch"
        ;;
      10)
        codename="buster"
        ;;
      11)
        codename="bullseye"
        ;;
      *)
        codename=""
        ;;
    esac
  fi

  echo "$codename"
}

collector_config_dir="/etc/otel/collector"
agent_config_path="${collector_config_dir}/agent_config.yaml"
gateway_config_path="${collector_config_dir}/gateway_config.yaml"
old_config_path="${collector_config_dir}/splunk_config_linux.yaml"
collector_env_path="${collector_config_dir}/splunk-otel-collector.conf"
collector_env_old_path="${collector_config_dir}/splunk_env"
collector_bundle_dir="/usr/lib/splunk-otel-collector/agent-bundle"
collectd_config_dir="${collector_bundle_dir}/run/collectd"
distro="$( get_distro )"
distro_codename="$( get_distro_codename )"
distro_version="$( get_distro_version )"
distro_arch="$( uname -m )"
repo_base="https://splunk.jfrog.io/splunk"

deb_repo_base="${repo_base}/otel-collector-deb"
debian_gpg_key_url="${deb_repo_base}/splunk-B3CD4420.gpg"

rpm_repo_base="${repo_base}/otel-collector-rpm"
yum_gpg_key_url="${rpm_repo_base}/splunk-B3CD4420.pub"

fluent_capng_c_version="0.2.2"
fluent_config_dir="${collector_config_dir}/fluentd"
fluent_config_path="${fluent_config_dir}/fluent.conf"
fluent_plugin_systemd_version="1.0.1"
journald_config_path="${fluent_config_dir}/conf.d/journald.conf"

td_agent_repo_base="https://packages.treasuredata.com"
td_agent_gpg_key_url="${td_agent_repo_base}/GPG-KEY-td-agent"

default_stage="release"
default_realm="us0"
default_memory_size="512"

default_collector_version="latest"
default_td_agent_version="4.3.2"
default_td_agent_version_stretch="3.7.1-0"

default_service_user="splunk-otel-collector"
default_service_group="splunk-otel-collector"

if [ "$distro_codename" = "stretch" ]; then
  default_td_agent_version="$default_td_agent_version_stretch"
fi

default_instrumentation_version="latest"
default_deployment_environment=""
instrumentation_config_path="/usr/lib/splunk-instrumentation/instrumentation.conf"
instrumentation_so_path="/usr/lib/splunk-instrumentation/libsplunk.so"
instrumentation_jar_path="/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"
generate_service_name="true"
service_name=""
disable_telemetry="false"
enable_profiler="false"
enable_profiler_memory="false"
enable_metrics="false"

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
  local version="${2:-}"

  if [ "$version" = "latest" ]; then
    version=""
  elif [ -n "$version" ]; then
    version="-${version}"
  fi

  if command -v yum >/dev/null 2>&1; then
    yum install -y ${package_name}${version}
  elif command -v dnf >/dev/null 2>&1; then
    dnf install -y ${package_name}${version}
  else
    zypper install -y ${package_name}${version}
  fi
}

ensure_not_installed() {
  local with_fluentd="$1"
  local with_instrumentation="$2"
  local agents="otelcol"

  if [ "$with_fluentd" = "true" ]; then
    agents="$agents td-agent"
  fi

  for agent in $agents; do
    if command -v $agent >/dev/null 2>&1; then
      echo "An agent binary already exists at $( command -v $agent ) which implies that the agent has already been installed." >&2
      echo "Please uninstall the agent and re-run this script." >&2
      exit 1
    fi
  done

  if [ "$with_instrumentation" = "true" ] && [ -f "$instrumentation_so_path" ]; then
    echo "$instrumentation_so_path already exists which implies that the instrumentation library has already been installed." >&2
    echo "Please uninstall the instrumentation library and re-run this script" >&2
    exit 1
  fi
}

configure_env_file() {
  local key="$1"
  local value="$2"
  local env_file="$3"

  echo "${key}=${value}" >> "$env_file"
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

fluent_plugin_installed() {
  local name="$1"

  td-agent-gem list "$name" --exact | grep -q "$name"
}

install_fluent_plugin() {
  local name="$1"
  local version="${2:-}"

  if [ -n "$version" ]; then
    td-agent-gem install "$name" --version "$version"
  else
    td-agent-gem install "$name"
  fi
}

configure_fluentd() {
  local override_src_path="$fluent_config_dir/splunk-otel-collector.conf"
  local override_dest_path="/etc/systemd/system/td-agent.service.d/splunk-otel-collector.conf"

  if [ -f "$override_src_path" ]; then
    systemctl stop td-agent
    mkdir -p $(dirname $override_dest_path)
    cp -f $override_src_path $override_dest_path
    chown root:root $override_dest_path
    chmod 644 $override_dest_path
    systemctl daemon-reload

    # ensure the td-agent user has access to the config dir
    chown -R td-agent:td-agent "$fluent_config_dir"

    # configure permissions/capabilities
    if [ -f /opt/td-agent/bin/fluent-cap-ctl ]; then
      if ! fluent_plugin_installed "capng_c"; then
        install_fluent_plugin "capng_c" "$fluent_capng_c_version"
      fi
      /opt/td-agent/bin/fluent-cap-ctl --add "dac_override,dac_read_search" -f /opt/td-agent/bin/ruby
    else
      if getent group adm >/dev/null 2>&1; then
        usermod -a -G adm td-agent
      fi
      if getent group systemd-journal 2>&1; then
        usermod -a -G systemd-journal td-agent
      fi
    fi

    if ! fluent_plugin_installed "fluent-plugin-systemd"; then
      install_fluent_plugin "fluent-plugin-systemd" "$fluent_plugin_systemd_version"
    fi
  fi
}

update_deployment_environment() {
  local deployment_environment="$1"

  if grep -q -E "^resource_attributes=.*deployment\.environment=${deployment_environment}(,|$)" "$instrumentation_config_path"; then
    echo "The 'deployment.environment=${deployment_environment}' resource attribute already exists in ${instrumentation_config_path}"
  else
    echo "Adding 'deployment.environment=${deployment_environment}' resource attribute to $instrumentation_config_path"
    if grep -q '^resource_attributes=' "$instrumentation_config_path"; then
      # traverse through existing resource attributes to add/update deployment.environment
      deployment_environment_found="false"
      attributes=""
      for i in $(grep '^resource_attributes=' "$instrumentation_config_path" | sed 's|^resource_attributes=||' | sed 's|,| |g'); do
        key="$(echo "$i" | cut -d= -f1)"
        value="$(echo "$i" | cut -d= -f2)"
        if [ "$key" = "deployment.environment" ]; then
          deployment_environment_found="true"
          value="$deployment_environment"
        fi
        attributes="${attributes},${key}=${value}"
      done
      if [ "$deployment_environment_found" != "true" ]; then
        attributes="${attributes},deployment.environment=${deployment_environment}"
      fi
      sed -i "s|^resource_attributes=.*|resource_attributes=${attributes#,}|" "$instrumentation_config_path"
    else
      # "resource_attributes=" line not found, simply append the line to the config file
      echo "resource_attributes=deployment.environment=${deployment_environment}" >> "$instrumentation_config_path"
    fi
  fi
}

update_instrumentation_option() {
  local option="$1"
  local value="$2"

  if grep -q -E "^${option}=.*" "$instrumentation_config_path"; then
    # overwrite existing option=value
    sed -i "s|^${option}=.*|${option}=${value}|" "$instrumentation_config_path"
  else
    # append option=value
    echo "${option}=${value}" >> "$instrumentation_config_path"
  fi
}

update_instrumentation_config() {
  local deployment_environment="$1"

  if [ -f "$instrumentation_config_path" ]; then
    ts="$(date '+%Y%m%d-%H%M%S')"
    echo "Backing up $instrumentation_config_path as ${instrumentation_config_path}.bak.${ts}"
    cp "$instrumentation_config_path" "${instrumentation_config_path}.bak.${ts}"
    if [ -n "$deployment_environment" ]; then
      update_deployment_environment "$deployment_environment"
    fi
    if [ -n "$service_name" ]; then
      update_instrumentation_option "service_name" "$service_name"
    fi
    update_instrumentation_option "generate_service_name" "$generate_service_name"
    update_instrumentation_option "disable_telemetry" "$disable_telemetry"
    update_instrumentation_option "enable_profiler" "$enable_profiler"
    update_instrumentation_option "enable_profiler_memory" "$enable_profiler_memory"
    update_instrumentation_option "enable_metrics" "$enable_metrics"
  else
    echo "$instrumentation_config_path not found!" >&2
    exit 1
  fi
}

install() {
  local stage="$1"
  local collector_version="$2"
  local td_agent_version="$3"
  local skip_collector_repo="$4"
  local skip_fluentd_repo="$5"
  local instrumentation_version="$6"

  case "$distro" in
    ubuntu|debian)
      if [ -z "$distro_codename" ]; then
        echo "The distribution codename could not be determined" >&2
        exit 1
      fi
      apt-get -y update
      apt-get -y install apt-transport-https gnupg
      if [ "$skip_collector_repo" = "false" ]; then
        install_collector_apt_repo "$stage"
      fi
      apt-get -y update
      install_apt_package "splunk-otel-collector" "$collector_version"
      if [ -n "$td_agent_version" ]; then
        if [ "$distro_codename" != "stretch" ]; then
          td_agent_version="${td_agent_version}-1"
        fi
        if [ "$skip_fluentd_repo" = "false" ]; then
          install_td_agent_apt_repo "$td_agent_version"
        fi
        apt-get -y update
        install_apt_package "td-agent" "$td_agent_version"
        apt-get -y install build-essential libcap-ng0 libcap-ng-dev pkg-config
        systemctl stop td-agent
      fi
      if [ -n "$instrumentation_version" ]; then
        install_apt_package "splunk-otel-auto-instrumentation" "$instrumentation_version"
      fi
      ;;
    amzn|centos|ol|rhel)
      if [ -z "$distro_version" ]; then
        echo "The distribution version could not be determined" >&2
        exit 1
      fi
      install_yum_package "libcap"
      if [ "$skip_collector_repo" = "false" ]; then
        install_collector_yum_repo "$stage"
      fi
      install_yum_package "splunk-otel-collector" "$collector_version"
      if [ -n "$td_agent_version" ]; then
        if [ "$skip_fluentd_repo" = "false" ]; then
          install_td_agent_yum_repo "$td_agent_version"
        fi
        install_yum_package "td-agent" "$td_agent_version"
        if command -v yum >/dev/null 2>&1; then
          yum group install -y 'Development Tools'
        else
          dnf group install -y 'Development Tools'
        fi
        for pkg in libcap-ng libcap-ng-devel pkgconfig; do
          install_yum_package "$pkg" ""
        done
        systemctl stop td-agent
      fi
      if [ -n "$instrumentation_version" ]; then
        install_yum_package "splunk-otel-auto-instrumentation" "$instrumentation_version"
      fi
      ;;
    sles|opensuse*)
      if [ "$skip_collector_repo" = "false" ]; then
        rpm --import $yum_gpg_key_url
        install_collector_yum_repo "$stage" "/etc/zypp/repos.d/"
      fi
      zypper -n --gpg-auto-import-keys refresh
      install_yum_package "libcap-progs"
      install_yum_package "splunk-otel-collector" "$collector_version"
      if [ -n "$instrumentation_version" ]; then
        install_yum_package "splunk-otel-auto-instrumentation" "$instrumentation_version"
      fi
      ;;
    *)
      echo "Your distro ($distro) is not supported or could not be determined" >&2
      exit 1
      ;;
  esac
}

uninstall() {
  for agent in otelcol td-agent $instrumentation_so_path; do
    if command -v $agent >/dev/null 2>&1; then
      pkg="$agent"
      if [ "$agent" = "otelcol" ]; then
        pkg="splunk-otel-collector"
      elif [ "$agent" = "$instrumentation_so_path" ]; then
        pkg="splunk-otel-auto-instrumentation"
      fi
      case "$distro" in
        ubuntu|debian)
          if dpkg -s $pkg >/dev/null 2>&1; then
            if [ "$pkg" != "splunk-otel-auto-instrumentation" ]; then
              systemctl stop $pkg || true
            fi
            apt-get purge -y $pkg 2>&1
            echo "Successfully removed the $pkg package"
          else
            agent_path="$( command -v agent )"
            echo "$agent_path exists but the $pkg package is not installed" >&2
            echo "$agent_path needs to be manually removed/uninstalled" >&2
            exit 1
          fi
          ;;
        amzn|centos|ol|rhel|sles|opensuse*)
          if rpm -q $pkg >/dev/null 2>&1; then
            if [ "$pkg" != "splunk-otel-auto-instrumentation" ]; then
              systemctl stop $pkg || true
            fi
            if command -v yum >/dev/null 2>&1; then
              yum remove -y $pkg 2>&1
            elif command -v dnf >/dev/null 2>&1; then
              dnf remove -y $pkg 2>&1
            else
              zypper remove -y $pkg
            fi
            echo "Successfully removed the $pkg package"
          else
            agent_path="$( command -v agent )"
            echo "$agent_path exists but the $pkg package is not installed" >&2
            echo "$agent_path needs to be manually removed/uninstalled" >&2
            exit 1
          fi
          ;;
        *)
          echo "Your distro ($distro) is not supported or could not be determined" >&2
          exit 1
          ;;
      esac
    fi
  done
}

usage() {
  cat <<EOH >&2
Usage: $0 [options] [access_token]

Installs the Splunk OpenTelemetry Collector for Linux from the package repos.
If access_token is not provided, it will be prompted for on stdin.

Options:

  --api-url <url>                       Set the api endpoint URL explicitly instead of the endpoint inferred from the
                                        specified realm.
                                        (default: https://api.REALM.signalfx.com)
  --ballast <ballast size>              Set the ballast size explicitly instead of the value calculated from the
                                        '--memory' option. This should be set to 1/3 to 1/2 of configured memory.
  --beta                                Use the beta package repo instead of the primary.
  --collector-config <path>             Set the path to an existing custom config file for the collector service instead
                                        of the default config file provided by the collector package based on the
                                        '--mode <agent|gateway>' option.
                                        *Note*: If the specified config file requires custom environment variables, the
                                        variables and values can be manually added to $collector_env_path
                                        after installation. Restart the collector service with the
                                        'sudo systemctl restart splunk-otel-collector' command for the changes to take
                                        effect.
  --collector-version <version>         The splunk-otel-collector package version to install.
                                        (default: "$default_collector_version")
  --hec-token <token>                   Set the HEC token if different than the specified access_token.
  --hec-url <url>                       Set the HEC endpoint URL explicitly instead of the endpoint inferred from the
                                        specified realm.
                                        (default: https://ingest.REALM.signalfx.com/v1/log)
  --ingest-url <url>                    Set the ingest endpoint URL explicitly instead of the endpoint inferred from the
                                        specified realm.
                                        (default: https://ingest.REALM.signalfx.com)
  --memory <memory size>                Total memory in MIB to allocate to the collector; automatically calculates the
                                        ballast size.
                                        (default: "$default_memory_size")
  --mode <agent|gateway>                Configure the collector service to run in agent or gateway mode.
                                        (default: "agent")
  --realm <us0|us1|eu0|...>             The Splunk realm to use. The ingest, api, trace, and HEC endpoint URLs will
                                        automatically be inferred by this value.
                                        (default: "$default_realm")
  --service-group <group>               Set the group for the splunk-otel-collector service. The group will be created
                                        if it does not exist.
                                        (default: "$default_service_group")
  --service-user <user>                 Set the user for the splunk-otel-collector service. The user will be created if
                                        it does not exist.
                                        (default: "$default_service_user")
  --skip-collector-repo                 By default, a apt/yum/zypper repo definition file will be created to download
                                        the collector deb/rpm package from $repo_base.
                                        Specify this option to skip this step and use a pre-configured repo on the
                                        target system that provides the 'splunk-otel-collector' deb/rpm package.
  --skip-fluentd-repo                   By default, a apt/yum repo definition file will be created to download the
                                        fluentd deb/rpm package from $td_agent_repo_base.
                                        Specify this option to skip this step and use a pre-configured repo on the
                                        target system that provides the 'td-agent' deb/rpm package.
  --test                                Use the test package repo instead of the primary.
  --trace-url <url>                     Set the trace endpoint URL explicitly instead of the endpoint inferred from the
                                        specified realm.
                                        (default: https://ingest.REALM.signalfx.com/v2/trace)
  --uninstall                           Removes the Splunk OpenTelemetry Collector for Linux.
  --with[out]-fluentd                   Whether to install and configure fluentd to forward log events to the collector.
                                        (default: --with-fluentd)
  --with[out]-instrumentation           Whether to install and configure the splunk-otel-auto-instrumentation package.
                                        (default: --without-instrumentation)
  --deployment-environment <value>      Set the 'deployment.environment' resource attribute to the specified value.
                                        Only applicable if the '--with-instrumentation' option is also specified.
                                        (default: empty)
  --service-name <name>                 Override the auto-generated service names for all instrumented Java applications
                                        on this host with '<name>'.
                                        Only applicable if the '--with-instrumentation' option is also specified.
                                        (default: empty)
  --[no-]generate-service-name          Specify '--no-generate-service-name' to prevent the preloader from setting the
                                        OTEL_SERVICE_NAME environment variable.
                                        Only applicable if the '--with-instrumentation' option is also specified.
                                        (default: --generate-service-name)
  --[enable|disable]-telemetry          Enable or disable the instrumentation preloader from sending the
                                        'splunk.linux-autoinstr.executions' metric to the collector.
                                        Only applicable if the '--with-instrumentation' option is also specified.
                                        (default: --enable-telemetry)
  --[enable|disable]-profiler           Enable or disable AlwaysOn CPU Profiling.
                                        Only applicable if the '--with-instrumentation' option is also specified.
                                        (default: --disable-profiler)
  --[enable|disable]-profiler-memory    Enable or disable AlwaysOn Memory Profiling.
                                        Only applicable if the '--with-instrumentation' option is also specified.
                                        (default: --disable-profiler-memory)
  --[enable|disable]-metrics            Enable or disable exporting Micrometer metrics.
                                        Only applicable if the '--with-instrumentation' option is also specified.
                                        (default: --disable-metrics)
  --instrumentation-version             The splunk-otel-auto-instrumentation package version to install.
                                        Only applicable if the '--with-instrumentation' option is also specified.
                                        (default: $default_instrumentation_version)
  --                                    Use '--' if access_token starts with '-'.

EOH
  exit 0
}

distro_is_supported() {
  case "$distro" in
    ubuntu)
      case "$distro_codename" in
        bionic|focal|xenial|jammy)
          return 0
          ;;
      esac
      ;;
    debian)
      case "$distro_codename" in
        bullseye|buster|stretch)
          return 0
          ;;
      esac
      ;;
    amzn)
      case "$distro_version" in
        2|2023)
          return 0
          ;;
      esac
      ;;
    sles|opensuse*)
      case "$distro_version" in
        12*|15*|42*)
          return 0
          ;;
      esac
      ;;
    centos|ol|rhel)
      case "$distro_version" in
        7*|8*|9*)
          return 0
          ;;
      esac
      ;;
  esac
  return 1
}

arch_supported() {
  case "$distro_arch" in
    amd64|x86_64|aarch64|arm64)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

fluentd_supported() {
  case "$distro" in
    amzn)
      if [ "$distro_version" != "2" ]; then
        return 1
      fi
      ;;
    sles|opensuse*)
      return 1
      ;;
    debian)
      if [ "$distro_version" = "9" ] && [ "$distro_arch" = "aarch64" ]; then
        return 1
      fi
      ;;
    ubuntu)
      if [ "$distro_version" = "16.04" ] && [ "$distro_arch" = "aarch64" ]; then
        return 1
      fi
      ;;
  esac

  return 0
}

check_support() {
  case "$distro" in
    debian|ubuntu)
      if [ -z "$distro_codename" ]; then
        echo "Your Linux distribution codename could not be determined from /etc/os-release." >&2
        exit 1
      fi
      ;;
    *)
      if [ -z "$distro" ]; then
        echo "Your Linux distribution could not be determined from /etc/os-release." >&2
        exit 1
      fi
      if [ -z "$distro_version" ]; then
        echo "Your Linux distribution version could not be determined from /etc/os-release." >&2
        exit 1
      fi
      if [ -z "$distro_arch" ]; then
        echo "Your system's architecture could not be determined from 'uname -m'." >&2
        exit 1
      fi
      ;;
  esac

  if ! distro_is_supported; then
    echo "Your Linux distribution/version is not supported." >&2
    exit 1
  fi

  if ! arch_supported; then
    echo "Your system's architecture '${distro_arch}' is not supported." >&2
    exit 1
  fi

  if ! command -v systemctl >/dev/null 2>&1; then
    echo "The systemctl command is required but was not found." >&2
    exit 1
  fi
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
  local mode="agent"
  local with_fluentd="true"
  local collector_config_path=
  local skip_collector_repo="false"
  local skip_fluentd_repo="false"
  local with_instrumentation="false"
  local instrumentation_version="$default_instrumentation_version"
  local deployment_environment="$default_deployment_environment"

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
      --collector-config)
        collector_config_path="$2"
        shift 1
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
      --mode)
        case $2 in
          agent|gateway)
            mode="$2"
            ;;
          *)
            echo "Unsupported mode '$2'" >&2
            exit 1
            ;;
        esac
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
      --skip-collector-repo)
        skip_collector_repo="true"
        ;;
      --skip-fluentd-repo)
        skip_fluentd_repo="true"
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
        if ! fluentd_supported; then
          echo "WARNING: Ignoring the --with-fluentd option since fluentd is currently not supported for ${distro}:${distro_version} ${distro_arch}." >&2
          with_fluentd="false"
        else
          with_fluentd="true"
        fi
        ;;
      --without-fluentd)
        with_fluentd="false"
        ;;
      --with-instrumentation)
        with_instrumentation="true"
        ;;
      --without-instrumentation)
        with_instrumentation="false"
        ;;
      --instrumentation-version)
        instrumentation_version="$2"
        shift 1
        ;;
      --deployment-environment)
        deployment_environment="$2"
        shift 1
        ;;
      --service-name)
        service_name="$2"
        shift 1
        ;;
      --generate-service-name)
        generate_service_name="true"
        ;;
      --no-generate-service-name)
        generate_service_name="false"
        ;;
      --enable-telemetry)
        disable_telemetry="false"
        ;;
      --disable-telemetry)
        disable_telemetry="true"
        ;;
      --enable-profiler)
        enable_profiler="true"
        ;;
      --disable-profiler)
        enable_profiler="false"
        ;;
      --enable-profiler-memory)
        enable_profiler_memory="true"
        ;;
      --disable-profiler-memory)
        enable_profiler_memory="false"
        ;;
      --enable-metrics)
        enable_metrics="true"
        ;;
      --disable-metrics)
        enable_metrics="false"
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

  if ! fluentd_supported; then
    with_fluentd="false"
  fi

  ensure_not_installed "$with_fluentd" "$with_instrumentation"

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

  if [ "$with_instrumentation" != "true" ]; then
    instrumentation_version=""
    deployment_environment=""
  fi

  if [ -z "$trace_url" ]; then
    trace_url="${ingest_url}/v2/trace"
  fi

  echo "Splunk OpenTelemetry Collector Version: ${collector_version}"
  if [ -n "$ballast" ]; then
    echo "Ballast Size in MIB: $ballast"
  fi
  echo "Memory Size in MIB: $memory"
  echo "Realm: $realm"
  echo "Ingest Endpoint: $ingest_url"
  echo "API Endpoint: $api_url"
  echo "Trace Endpoint: $trace_url"
  echo "HEC Endpoint: $hec_url"
  if [ "$with_fluentd" = "true" ]; then
    echo "TD Agent (Fluentd) Version: $td_agent_version"
  fi
  if [ "$with_instrumentation" = "true" ]; then
    echo "Splunk OpenTelemetry Auto Instrumentation Version: $instrumentation_version"
    if [ -n "$deployment_environment" ]; then
      echo "  Resource Attribute: deployment.environment=${deployment_environment}"
    fi
  fi

  if [ "${VERIFY_ACCESS_TOKEN:-true}" = "true" ] && ! verify_access_token "$access_token" "$ingest_url" "$insecure"; then
    echo "Your access token could not be verified. This may be due to a network connectivity issue or an invalid access token." >&2
    exit 1
  fi

  install "$stage" "$collector_version" "$td_agent_version" "$skip_collector_repo" "$skip_fluentd_repo" "$instrumentation_version"

  if [ "$with_instrumentation" = "true" ]; then
    update_instrumentation_config "$deployment_environment"
  fi

  create_user_group "$service_user" "$service_group"
  configure_service_owner "$service_user" "$service_group"

  if [ -z "$collector_config_path" ]; then
    # custom config not provided; use the config provided by the collector package based on the --mode option
    if [ "$mode" = "agent" ]; then
      if [ -f "$agent_config_path" ]; then
        # use the agent config if the installed package includes it
        collector_config_path="$agent_config_path"
      elif [ -f "$old_config_path" ]; then
        # use the old config if the installed package does not include the new agent config
        collector_config_path="$old_config_path"
      fi
    else
      if [ -f "$gateway_config_path" ]; then
        # use the gateway config if the installed package includes it
        collector_config_path="$gateway_config_path"
      elif [ -f "$agent_config_path" ]; then
        # use the agent config if the installed package includes it
        collector_config_path="$agent_config_path"
      elif [ -f "$old_config_path" ]; then
        # use the old config if the installed package does not include the new agent or gateway config
        collector_config_path="$old_config_path"
      fi
    fi
  fi

  if [ -z "$collector_config_path" ]; then
    echo "ERROR: The installed splunk-otel-collector package does not include a supported config file!" >&2
    exit 1
  elif [ ! -f "$collector_config_path" ]; then
    echo "ERROR: Config file $collector_config_path not found!" >&2
    exit 1
  fi

  if [ ! -f "${collector_env_path}.example" ]; then
    collector_env_path=$collector_env_old_path
  fi

  mkdir -p "$(dirname $collector_env_path)"

  # remove existing env file and recreate with current values
  if [ -f "$collector_env_path" ]; then
    rm -f "$collector_env_path"
  fi

  configure_env_file "SPLUNK_CONFIG" "$collector_config_path" "$collector_env_path"
  configure_env_file "SPLUNK_ACCESS_TOKEN" "$access_token" "$collector_env_path"
  configure_env_file "SPLUNK_REALM" "$realm" "$collector_env_path"
  configure_env_file "SPLUNK_API_URL" "$api_url" "$collector_env_path"
  configure_env_file "SPLUNK_INGEST_URL" "$ingest_url" "$collector_env_path"
  configure_env_file "SPLUNK_TRACE_URL" "$trace_url" "$collector_env_path"
  configure_env_file "SPLUNK_HEC_URL" "$hec_url" "$collector_env_path"
  configure_env_file "SPLUNK_HEC_TOKEN" "$hec_token" "$collector_env_path"
  configure_env_file "SPLUNK_MEMORY_TOTAL_MIB" "$memory" "$collector_env_path"
  if [ -n "$ballast" ]; then
    configure_env_file "SPLUNK_BALLAST_SIZE_MIB" "$ballast" "$collector_env_path"
  fi
  if [ -d "$collector_bundle_dir" ]; then
    configure_env_file "SPLUNK_BUNDLE_DIR" "$collector_bundle_dir" "$collector_env_path"
    # ensure the collector service owner has access to the bundle dir
    chown -R $service_user:$service_group "$(dirname $collector_bundle_dir)"
  fi
  if [ -d "$collectd_config_dir" ]; then
    configure_env_file "SPLUNK_COLLECTD_DIR" "$collectd_config_dir" "$collector_env_path"
    # ensure the collector service owner has access to the collectd dir
    chown -R $service_user:$service_group "$(dirname $collectd_config_dir)"
  fi

  # ensure the collector service owner has access to the config dir
  chown -R $service_user:$service_group "$(dirname $collector_config_dir)"

  # ensure only the collector service owner has access to the env file
  chmod 600 "$collector_env_path"

  # delete the default user/group if a custom service user/group was specified
  if [ "$service_user" != "$default_service_user" ] && getent passwd "$default_service_user" >/dev/null 2>&1; then
    userdel "$default_service_user"
  fi
  if [ "$service_group" != "$default_service_group" ] && getent group "$default_service_group" >/dev/null 2>&1; then
    groupdel "$default_service_group"
  fi

  systemctl daemon-reload
  systemctl restart splunk-otel-collector

  if [ "$with_fluentd" = "true" ]; then
    # only start fluentd with our custom config to avoid port conflicts within the default config
    systemctl stop td-agent
    if [ -f "$fluent_config_path" ]; then
      configure_fluentd
      systemctl restart td-agent
    else
      if [ -f /etc/td-agent/td-agent.conf ]; then
        mv -f /etc/td-agent/td-agent.conf /etc/td-agent/td-agent.conf.bak
      fi
      systemctl disable td-agent
    fi
  fi

  cat <<EOH
The Splunk OpenTelemetry Collector for Linux has been successfully installed.

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

By default, fluentd has been configured to collect systemd journal log events from /var/log/journal.
See $journald_config_path for the default source configuration.

If the fluentd configuration is modified or new config files are added, the fluentd service must be
restarted to apply the changes by running the following command as root:

  systemctl restart td-agent

EOH
  fi

  if [ "$with_instrumentation" = "true" ]; then
    cat <<EOH
The Splunk OpenTelemetry Auto Instrumentation package has been installed.
/etc/ld.so.preload has been configured for the instrumentation library at $instrumentation_so_path.
The configuration file is located at $instrumentation_config_path.

The Java application(s) on the host need to be manually started/restarted.

EOH
  fi
  exit 0
}

check_support

parse_args_and_install $@
