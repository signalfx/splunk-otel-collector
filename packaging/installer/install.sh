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
      11)
        codename="bullseye"
        ;;
      12)
        codename="bookworm"
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
default_listen_interface="0.0.0.0"

default_collector_version="latest"
default_td_agent_version="4.3.2"

default_service_user="splunk-otel-collector"
default_service_group="splunk-otel-collector"

preload_path="/etc/ld.so.preload"
default_instrumentation_version="latest"
default_deployment_environment=""
instrumentation_so_path="/usr/lib/splunk-instrumentation/libsplunk.so"
instrumentation_jar_path="/usr/lib/splunk-instrumentation/splunk-otel-javaagent.jar"
systemd_instrumentation_config_path="/usr/lib/systemd/system.conf.d/00-splunk-otel-auto-instrumentation.conf"
service_name=""
enable_profiler="false"
enable_profiler_memory="false"
enable_metrics="false"
otlp_endpoint=""
otlp_endpoint_protocol=""
metrics_exporter=""
logs_exporter=""
java_zeroconfig_path="/etc/splunk/zeroconfig/java.conf"
node_zeroconfig_path="/etc/splunk/zeroconfig/node.conf"
dotnet_zeroconfig_path="/etc/splunk/zeroconfig/dotnet.conf"
node_package_path="/usr/lib/splunk-instrumentation/splunk-otel-js.tgz"
node_install_prefix="/usr/lib/splunk-instrumentation/splunk-otel-js"
dotnet_install_dir="/usr/lib/splunk-instrumentation/splunk-otel-dotnet"
dotnet_agent_path="${dotnet_install_dir}/linux-x64/OpenTelemetry.AutoInstrumentation.Native.so"
all_sdks="java node dotnet"
sdks_to_enable=""
sdks_enabled=""

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
  elif [ -f "$version" ]; then
    dpkg -i "$version"
    return $?
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
  elif [ -f "$version" ]; then
    rpm -Uvh "$version"
    return $?
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
  local with_systemd_instrumentation="$3"
  local npm_path="$4"
  local otelcol_path=$( command -v otelcol 2>/dev/null || true )
  local td_agent_path=$( command -v td-agent 2>/dev/null || true )

  if [ -n "$otelcol_path" ]; then
    echo "$otelcol_path already exists which implies that the collector is already installed." >&2
    echo "Please uninstall the collector, or try running this script with the '--uninstall' option." >&2
    exit 1
  fi

  if [ "$with_fluentd" = "true" ] && [ -n "$td_agent_path" ]; then
    echo "$td_agent_path already exists which implies that fluentd/td-agent is already installed." >&2
    echo "Please uninstall fluentd/td-agent, or try running this script with the '--uninstall' option." >&2
    exit 1
  fi

  if [ "$with_instrumentation" = "true" ] || [ "$with_systemd_instrumentation" = "true" ]; then
    if [ -f "$instrumentation_so_path" ]; then
      echo "$instrumentation_so_path already exists which implies that auto instrumentation is already installed." >&2
      echo "Please uninstall auto instrumentation, or try running this script with the '--uninstall' option." >&2
      exit 1
    fi
    if [ -f "$systemd_instrumentation_config_path" ]; then
      echo "$systemd_instrumentation_config_path already exists which implies that auto instrumentation is already installed." >&2
      echo "Please uninstall auto instrumentation, or try running this script with the '--uninstall' option." >&2
      exit 1
    fi
    if splunk_otel_js_installed "$npm_path"; then
      echo "The @splunk/otel npm package is already installed in $node_install_prefix." >&2
      echo "Please uninstall @splunk/otel, or try running this script with the '--uninstall' option." >&2
      exit 1
    fi
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

backup_file() {
  local path="$1"

  if [ -f "$path" ]; then
    ts="$(date '+%Y%m%d-%H%M%S')"
    echo "Backing up $path as ${path}.bak.${ts}"
    cp "$path" "${path}.bak.${ts}"
  fi
}

get_package_version() {
  local package="$1"
  local version=""

  case "$distro" in
    ubuntu|debian)
      version="$( dpkg-query --showformat='${Version}' --show $package )"
      ;;
    *)
      version="$( rpm -q --queryformat='%{VERSION}' $package )"
      ;;
  esac

  echo -n "$version" | sed 's|~|-|'
}

enable_preload() {
  if [ -f "$preload_path" ]; then
    if ! grep -q "$instrumentation_so_path" "$preload_path"; then
      backup_file "$preload_path"
      echo "Adding $instrumentation_so_path to $preload_path"
      echo "$instrumentation_so_path" >> "$preload_path"
    fi
  else
    echo "Adding $instrumentation_so_path to $preload_path"
    echo "$instrumentation_so_path" >> "$preload_path"
  fi
}

disable_preload() {
  if [ -f "$preload_path" ] && grep -q "$instrumentation_so_path" "$preload_path"; then
    backup_file "$preload_path"
    echo "Removing ${instrumentation_so_path} from ${preload_path}"
    sed -i -e "s|$instrumentation_so_path||" "$preload_path"
    if [ ! -s "$preload_path" ] || ! grep -q '[^[:space:]]' "$preload_path"; then
      echo "Removing empty ${preload_path}"
      rm -f "$preload_path"
    fi
  fi
}

create_zeroconfig_java() {
  local version="$( get_package_version splunk-otel-auto-instrumentation )"
  local resource_attributes="splunk.zc.method=splunk-otel-auto-instrumentation-${version}"

  if [ -n "$deployment_environment" ]; then
    resource_attributes="${resource_attributes},deployment.environment=${deployment_environment}"
  fi

  backup_file "$java_zeroconfig_path"

  echo "Creating ${java_zeroconfig_path}"
  cat <<EOH > $java_zeroconfig_path
JAVA_TOOL_OPTIONS=-javaagent:${instrumentation_jar_path}
OTEL_RESOURCE_ATTRIBUTES=${resource_attributes}
SPLUNK_PROFILER_ENABLED=${enable_profiler}
SPLUNK_PROFILER_MEMORY_ENABLED=${enable_profiler_memory}
SPLUNK_METRICS_ENABLED=${enable_metrics}
EOH

  if [ -n "$service_name" ]; then
    echo "OTEL_SERVICE_NAME=${service_name}" >> $java_zeroconfig_path
  fi

  if [ -n "$otlp_endpoint" ]; then
    echo "OTEL_EXPORTER_OTLP_ENDPOINT=${otlp_endpoint}" >> $java_zeroconfig_path
  fi

  if [ -n "$otlp_endpoint_protocol" ]; then
    echo "OTEL_EXPORTER_OTLP_PROTOCOL=${otlp_endpoint_protocol}" >> $java_zeroconfig_path
  fi

  if [ -n "$metrics_exporter" ]; then
    echo "OTEL_METRICS_EXPORTER=${metrics_exporter}" >> $java_zeroconfig_path
  fi

  if [ -n "$logs_exporter" ]; then
    echo "OTEL_LOGS_EXPORTER=${logs_exporter}" >> $java_zeroconfig_path
  fi
}

splunk_otel_js_installed() {
  local npm_path="$1"

  command -v "$npm_path" >/dev/null 2>&1 && [ -d "$node_install_prefix" ] && (cd "$node_install_prefix" && "$npm_path" ls --global=false @splunk/otel >/dev/null 2>&1)
}

install_node_package() {
  local npm_path="$1"

  if ! command -v "$npm_path" >/dev/null 2>&1; then
    return 1
  fi

  if [ "$distro_arch" = "arm64" ] || [ "$distro_arch" = "aarch64" ]; then
    echo "Installing dependencies for the Node.js Auto Instrumentation package ..."
    case "$distro" in
      ubuntu|debian)
        apt-get install -y build-essential
        ;;
      amzn|centos|ol|rhel|rocky)
        if command -v yum >/dev/null 2>&1; then
          yum group install -y 'Development Tools'
        else
          dnf group install -y 'Development Tools'
        fi
        ;;
      sles|opensuse*)
        zypper -n install -t pattern devel_basis
        zypper -n install -t pattern devel_C_C++
        ;;
    esac
  fi

  echo "Installing the Node.js Auto Instrumentation package ..."
  mkdir -p ${node_install_prefix}/node_modules
  echo "Running 'cd $node_install_prefix && $npm_path install --global=false $node_package_path':"
  (cd $node_install_prefix && $npm_path install --global=false $node_package_path)
}

create_zeroconfig_node() {
  local version="$( get_package_version splunk-otel-auto-instrumentation )"
  local resource_attributes="splunk.zc.method=splunk-otel-auto-instrumentation-${version}"

  if [ -n "$deployment_environment" ]; then
    resource_attributes="${resource_attributes},deployment.environment=${deployment_environment}"
  fi

  backup_file "$node_zeroconfig_path"

  echo "Creating ${node_zeroconfig_path}"
  cat <<EOH > $node_zeroconfig_path
NODE_OPTIONS=-r ${node_install_prefix}/node_modules/@splunk/otel/instrument
OTEL_RESOURCE_ATTRIBUTES=${resource_attributes}
SPLUNK_PROFILER_ENABLED=${enable_profiler}
SPLUNK_PROFILER_MEMORY_ENABLED=${enable_profiler_memory}
SPLUNK_METRICS_ENABLED=${enable_metrics}
EOH

  if [ -n "$service_name" ]; then
    echo "OTEL_SERVICE_NAME=${service_name}" >> $node_zeroconfig_path
  fi

  if [ -n "$otlp_endpoint" ]; then
    echo "OTEL_EXPORTER_OTLP_ENDPOINT=${otlp_endpoint}" >> $node_zeroconfig_path
  fi

  if [ -n "$otlp_endpoint_protocol" ]; then
    echo "OTEL_EXPORTER_OTLP_PROTOCOL=${otlp_endpoint_protocol}" >> $node_zeroconfig_path
  fi

  if [ -n "$metrics_exporter" ]; then
    echo "OTEL_METRICS_EXPORTER=${metrics_exporter}" >> $node_zeroconfig_path
  fi

  if [ -n "$logs_exporter" ]; then
    echo "OTEL_LOGS_EXPORTER=${logs_exporter}" >> $node_zeroconfig_path
  fi
}

create_zeroconfig_dotnet() {
  local version="$( get_package_version splunk-otel-auto-instrumentation )"
  local resource_attributes="splunk.zc.method=splunk-otel-auto-instrumentation-${version}"

  if [ -n "$deployment_environment" ]; then
    resource_attributes="${resource_attributes},deployment.environment=${deployment_environment}"
  fi

  backup_file "$dotnet_zeroconfig_path"

  echo "Creating ${dotnet_zeroconfig_path}"
  cat <<EOH > $dotnet_zeroconfig_path
CORECLR_ENABLE_PROFILING=1
CORECLR_PROFILER={918728DD-259F-4A6A-AC2B-B85E1B658318}
CORECLR_PROFILER_PATH=${dotnet_install_dir}/linux-x64/OpenTelemetry.AutoInstrumentation.Native.so
DOTNET_ADDITIONAL_DEPS=${dotnet_install_dir}/AdditionalDeps
DOTNET_SHARED_STORE=${dotnet_install_dir}/store
DOTNET_STARTUP_HOOKS=${dotnet_install_dir}/net/OpenTelemetry.AutoInstrumentation.StartupHook.dll
OTEL_DOTNET_AUTO_HOME=${dotnet_install_dir}
OTEL_DOTNET_AUTO_PLUGINS=Splunk.OpenTelemetry.AutoInstrumentation.Plugin,Splunk.OpenTelemetry.AutoInstrumentation
OTEL_RESOURCE_ATTRIBUTES=${resource_attributes}
SPLUNK_PROFILER_ENABLED=${enable_profiler}
SPLUNK_PROFILER_MEMORY_ENABLED=${enable_profiler_memory}
SPLUNK_METRICS_ENABLED=${enable_metrics}
EOH

  if [ -n "$service_name" ]; then
    echo "OTEL_SERVICE_NAME=${service_name}" >> $dotnet_zeroconfig_path
  fi

  if [ -n "$otlp_endpoint" ]; then
    echo "OTEL_EXPORTER_OTLP_ENDPOINT=${otlp_endpoint}" >> $dotnet_zeroconfig_path
  fi

  if [ -n "$otlp_endpoint_protocol" ]; then
    echo "OTEL_EXPORTER_OTLP_PROTOCOL=${otlp_endpoint_protocol}" >> $dotnet_zeroconfig_path
  fi

  if [ -n "$metrics_exporter" ]; then
    echo "OTEL_METRICS_EXPORTER=${metrics_exporter}" >> $dotnet_zeroconfig_path
  fi

  if [ -n "$logs_exporter" ]; then
    echo "OTEL_LOGS_EXPORTER=${logs_exporter}" >> $dotnet_zeroconfig_path
  fi
}

create_systemd_instrumentation_config() {
  local sdks="$1"
  local version="$( get_package_version splunk-otel-auto-instrumentation )"
  local resource_attributes="splunk.zc.method=splunk-otel-auto-instrumentation-${version}-systemd"

  if [ -n "$deployment_environment" ]; then
    resource_attributes="${resource_attributes},deployment.environment=${deployment_environment}"
  fi

  mkdir -p "$(dirname $systemd_instrumentation_config_path)"

  backup_file "$systemd_instrumentation_config_path"

  echo "Creating ${systemd_instrumentation_config_path}"
  cat <<EOH > $systemd_instrumentation_config_path
[Manager]
DefaultEnvironment="OTEL_RESOURCE_ATTRIBUTES=${resource_attributes}"
DefaultEnvironment="SPLUNK_PROFILER_ENABLED=${enable_profiler}"
DefaultEnvironment="SPLUNK_PROFILER_MEMORY_ENABLED=${enable_profiler_memory}"
DefaultEnvironment="SPLUNK_METRICS_ENABLED=${enable_metrics}"
EOH

  if [ -n "$service_name" ]; then
    echo "DefaultEnvironment=\"OTEL_SERVICE_NAME=${service_name}\"" >> $systemd_instrumentation_config_path
  fi

  if [ -n "$otlp_endpoint" ]; then
    echo "DefaultEnvironment=\"OTEL_EXPORTER_OTLP_ENDPOINT=${otlp_endpoint}\"" >> $systemd_instrumentation_config_path
  fi

  if [ -n "$otlp_endpoint_protocol" ]; then
    echo "DefaultEnvironment=\"OTEL_EXPORTER_OTLP_PROTOCOL=${otlp_endpoint_protocol}\"" >> $systemd_instrumentation_config_path
  fi

  if [ -n "$metrics_exporter" ]; then
    echo "DefaultEnvironment=\"OTEL_METRICS_EXPORTER=${metrics_exporter}\"" >> $systemd_instrumentation_config_path
  fi

  if [ -n "$logs_exporter" ]; then
    echo "DefaultEnvironment=\"OTEL_LOGS_EXPORTER=${logs_exporter}\"" >> $systemd_instrumentation_config_path
  fi

  if item_in_list "java" "$sdks"; then
    echo "DefaultEnvironment=\"JAVA_TOOL_OPTIONS=-javaagent:${instrumentation_jar_path}\"" >> $systemd_instrumentation_config_path
  fi

  if item_in_list "node" "$sdks"; then
    echo "DefaultEnvironment=\"NODE_OPTIONS=-r ${node_install_prefix}/node_modules/@splunk/otel/instrument\"" >> $systemd_instrumentation_config_path
  fi

  if item_in_list "dotnet" "$sdks"; then
      cat <<EOH >> $systemd_instrumentation_config_path
DefaultEnvironment="CORECLR_ENABLE_PROFILING=1"
DefaultEnvironment="CORECLR_PROFILER={918728DD-259F-4A6A-AC2B-B85E1B658318}"
DefaultEnvironment="CORECLR_PROFILER_PATH=${dotnet_install_dir}/linux-x64/OpenTelemetry.AutoInstrumentation.Native.so"
DefaultEnvironment="DOTNET_ADDITIONAL_DEPS=${dotnet_install_dir}/AdditionalDeps"
DefaultEnvironment="DOTNET_SHARED_STORE=${dotnet_install_dir}/store"
DefaultEnvironment="DOTNET_STARTUP_HOOKS=${dotnet_install_dir}/net/OpenTelemetry.AutoInstrumentation.StartupHook.dll"
DefaultEnvironment="OTEL_DOTNET_AUTO_HOME=${dotnet_install_dir}"
DefaultEnvironment="OTEL_DOTNET_AUTO_PLUGINS=Splunk.OpenTelemetry.AutoInstrumentation.Plugin,Splunk.OpenTelemetry.AutoInstrumentation"
EOH
  fi

  systemctl daemon-reload
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
        td_agent_version="${td_agent_version}-1"
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
    amzn|centos|ol|rhel|rocky)
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
  local npm_path="$1"

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
            if [ "$pkg" = "splunk-otel-auto-instrumentation" ] && [ -f "$systemd_instrumentation_config_path" ]; then
              backup_file "$systemd_instrumentation_config_path"
              echo "Removing ${systemd_instrumentation_config_path}"
              rm -f "$systemd_instrumentation_config_path"
              systemctl daemon-reload
            fi
          else
            agent_path="$( command -v agent )"
            echo "$agent_path exists but the $pkg package is not installed" >&2
            echo "$agent_path needs to be manually removed/uninstalled" >&2
            exit 1
          fi
          ;;
        amzn|centos|ol|rhel|rocky|sles|opensuse*)
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
            if [ "$pkg" = "splunk-otel-auto-instrumentation" ] && [ -f "$systemd_instrumentation_config_path" ]; then
              backup_file "$systemd_instrumentation_config_path"
              echo "Removing ${systemd_instrumentation_config_path}"
              rm -f "$systemd_instrumentation_config_path"
              systemctl daemon-reload
            fi
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

  if splunk_otel_js_installed "$npm_path"; then
    (cd $node_install_prefix && "$npm_path" uninstall --global=false @splunk/otel)
    echo "Successfully uninstalled the @splunk/otel npm package from $node_install_prefix"
  fi
}

usage() {
  cat <<EOH >&2
Usage: $0 [options] [access_token]

Installs the Splunk OpenTelemetry Collector for Linux from the package repos.
If access_token is not provided, it will be prompted for on stdin.

Options:

Collector:
  -- <access_token>                     Use '--' if access_token starts with '-'.
  --api-url <url>                       Set the api endpoint URL explicitly instead of the endpoint inferred from the
                                        specified realm.
                                        (default: https://api.REALM.signalfx.com)
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
  --discovery                           Enable discovery mode on collector startup (disabled by default).
  --hec-token <token>                   Set the HEC token if different than the specified access_token.
  --hec-url <url>                       Set the HEC endpoint URL explicitly instead of the endpoint inferred from the
                                        specified realm.
                                        (default: https://ingest.REALM.signalfx.com/v1/log)
  --ingest-url <url>                    Set the ingest endpoint URL explicitly instead of the endpoint inferred from the
                                        specified realm.
                                        (default: https://ingest.REALM.signalfx.com)
  --memory <memory size>                Total memory in MIB to allocate to the collector
                                        (default: "$default_memory_size")
  --mode <agent|gateway>                Configure the collector service to run in agent or gateway mode.
                                        (default: "agent")
  --listen-interface <ip>               network interface the collector receivers listen on.
                                        (default: "127.0.0.1" for agent mode and "0.0.0.0" otherwise)
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
  --test                                Use the test package repo instead of the primary.

Fluentd [DEPRECATED]:
  --with[out]-fluentd                   Whether to install and configure fluentd to forward log events to the collector.
                                        (default: --without-fluentd)
  --skip-fluentd-repo                   By default, a apt/yum repo definition file will be created to download the
                                        fluentd deb/rpm package from $td_agent_repo_base.
                                        Specify this option to skip this step and use a pre-configured repo on the
                                        target system that provides the 'td-agent' deb/rpm package.
                                        Only applicable if the '--with-fluentd' is also specified.

Auto Instrumentation:
  --with[out]-instrumentation           Whether to install the splunk-otel-auto-instrumentation package and add the
                                        libsplunk.so shared object library to /etc/ld.so.preload to enable auto
                                        instrumentation for all supported processes on the host.
                                        Cannot be combined with the '--with-systemd-instrumentation' option.
                                        (default: --without-instrumentation)
  --with[out]-systemd-instrumentation   Whether to install the splunk-otel-auto-instrumentation package and configure a
                                        systemd drop-in file to enable auto instrumentation for all supported
                                        applications running as systemd services.
                                        Cannot be combined with the '--with-instrumentation' option.
                                        (default: --without-systemd-instrumentation)
  --with[out]-instrumentation-sdk "<s>" Whether to enable Auto Instrumentation for a specific language. This option
                                        takes a comma separated set of values representing supported
                                        auto-instrumentation SDKs.
                                        Currently supported values: "java", "node", and "dotnet"
                                        Use --with-instrumentation-sdk to enable only the specified language(s),
                                        for example "--with-instrumentation-sdk java".
                                        *Note*: .NET (dotnet) auto instrumentation is only supported on x86_64/amd64.
                                        (default: --with-instrumentation-sdk $( echo $all_sdks | tr ' ' ',' ))
  --npm-path <path>                     If Auto Instrumentation for Node.js is enabled, npm is required to install the
                                        included Splunk OpenTelemetry Auto Instrumentation for Node.js package. If npm
                                        is not found via the "command -v npm" shell command or if installation fails,
                                        Auto Instrumentation for Node.js will not be activated. Use this option to
                                        specify a custom path to npm, for example "/my/path/to/npm".
                                        (default: npm)
  --deployment-environment <value>      Set the "deployment.environment" resource attribute to the specified value.
                                        If not specified, the "Environment" in the Splunk APM UI will appear as
                                        "unknown" for all instrumented applications. The resource attribute will be
                                        appended to the OTEL_RESOURCE_ATTRIBUTES environment variable.
                                        (default: empty)
  --service-name <name>                 Override the auto-generated service names for all instrumented applications
                                        on this host with the specified value. The value will be set to the
                                        OTEL_SERVICE_NAME environment variable.
                                        (default: empty)
  --otlp-endpoint <host:port>           Set the OTLP endpoint for captured traces, logs, and metrics for all activated
                                        SDKs. The value will be set to the OTEL_EXPORTER_OTLP_ENDPOINT environment
                                        variable.
                                        (default: empty, i.e. defer to the default OTEL_EXPORTER_OTLP_ENDPOINT value for
                                        each activated SDK)
  --otlp-endpoint-protocol <protocol>   Set the protocol for the OTLP endpoint, for example "grpc" or "http/protobuf".
                                        The value will be set to the OTEL_EXPORTER_OTLP_PROTOCOL environment variable.
                                        Only applicable if the "--otlp-endpoint <host:port>" option is also specified.
                                        (default: empty, i.e. defer to the default OTEL_EXPORTER_OTLP_PROTOCOL value for
                                        each activated SDK)
  --metrics-exporter <exporters>        Comma-separated list of exporters for collected metrics by all activated SDKs,
                                        for example "otlp,prometheus". Set the value to "none" to disable collection and
                                        export of metrics. The value will be set to the OTEL_METRICS_EXPORTER
                                        environment variable.
                                        (default: empty, i.e. defer to the default OTEL_METRICS_EXPORTER value for each
                                        activated SDK)
  --logs-exporter <exporter>            Set the exporter for collected logs by all activated SDKs, for example "otlp".
                                        Set the value to "none" to disable collection and export of logs. The value will
                                        be set to the OTEL_LOGS_EXPORTER environment variable.
                                        (default: empty, i.e. defer to the default OTEL_LOGS_EXPORTER value for each
                                        activated SDK)
  --[enable|disable]-profiler           Enable or disable AlwaysOn Profiling for all activated SDKs that support the
                                        SPLUNK_PROFILER_ENABLED environment variable.
                                        (default: --disable-profiler)
  --[enable|disable]-profiler-memory    Enable or disable AlwaysOn Memory Profiling for all activated SDKs that support
                                        the SPLUNK_PROFILER_MEMORY_ENABLED environment variable.
                                        (default: --disable-profiler-memory)
  --[enable|disable]-metrics            Enable or disable instrumentation metrics collection for all activated SDKs that
                                        support the SPLUNK_METRICS_ENABLED environment variable.
                                        (default: --disable-metrics)
  --instrumentation-version             The splunk-otel-auto-instrumentation package version to install.
                                        *Note*: The minimum supported version for Java and Node.js auto instrumentation
                                        is 0.87.0, and the minimum supported version for .NET auto instrumentation is
                                        0.99.0.
                                        (default: $default_instrumentation_version)

Uninstall:
  --uninstall                           Removes the Splunk OpenTelemetry Collector for Linux, Fluentd, and Splunk
                                        OpenTelemetry Auto Instrumentation packages, if installed.

EOH
}

distro_is_supported() {
  case "$distro" in
    ubuntu)
      case "$distro_codename" in
        bionic|focal|xenial|jammy|noble)
          return 0
          ;;
      esac
      ;;
    debian)
      case "$distro_codename" in
        bookworm|bullseye)
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
    centos|ol|rhel|rocky)
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
      elif [ "$distro_version" = "12" ]; then
        return 1
      fi
      ;;
    ubuntu)
      if [ "$distro_version" = "16.04" ] && [ "$distro_arch" = "aarch64" ]; then
        return 1
      fi
      if [ "$distro_version" = "24.04" ]; then
        return 1
      fi
      ;;
  esac

  return 0
}

version_supported() {
  local min="$1"
  local desired="$2"
  local stage="$3"

  if [ "$min" = "$desired" ] || [ "$desired" = "latest" ] || [ "$stage" != "release" ] || [ -f "$desired" ]; then
    return 0
  fi

  if ! echo "$desired" | grep -q "^[[:digit:]]\+\.[[:digit:]]\+\.[[:digit:]]\+$"; then
    echo "[ERROR] Unsupported version: $desired" >&2
    exit 1
  fi

  for field in 1 2 3; do
    m=$( echo "$min" | cut -d "." -f $field )
    d=$( echo "$desired" | cut -d "." -f $field )
    if [ $d -lt $m ]; then
      return 1
    elif [ $d -gt $m ]; then
      return 0
    fi
  done
}

dotnet_supported() {
  case "$distro_arch" in
    amd64|x86_64)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

item_in_list() {
  local item="$1"
  shift
  local list="$@"

  for i in $list; do
    if [ "$i" = "$item" ]; then
      return 0
    fi
  done

  return 1
}

add_item_to_list() {
  local item="$1"
  shift
  local list="$@"

  if item_in_list "$list" "$item"; then
    echo $list
  else
    echo $list $item
  fi
}

remove_item_from_list() {
  local item="$1"
  shift
  local list="$@"
  local new_list=""

  for i in $list; do
    if [ "$i" != "$item" ]; then
      new_list="$new_list $i"
    fi
  done

  echo $new_list
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
  local collector_version="$default_collector_version"
  local hec_token=
  local hec_url=
  local ingest_url=
  local insecure=
  local memory="$default_memory_size"
  local listen_interface=
  local realm="$default_realm"
  local service_group="$default_service_group"
  local stage="$default_stage"
  local service_user="$default_service_user"
  local td_agent_version="$default_td_agent_version"
  local uninstall="false"
  local mode="agent"
  local with_fluentd="false"
  local collector_config_path=
  local skip_collector_repo="false"
  local skip_fluentd_repo="false"
  local with_instrumentation="false"
  local with_systemd_instrumentation="false"
  local instrumentation_version="$default_instrumentation_version"
  local deployment_environment="$default_deployment_environment"
  local discovery=
  local npm_path="npm"
  local node_package_installed="false"
  local with_sdks=""
  local without_sdks=""

  while [ -n "${1-}" ]; do
    case $1 in
      --api-url)
        api_url="$2"
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
      --listen-interface)
        listen_interface="$2"
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
      --uninstall)
        uninstall="true"
        ;;
      --with-fluentd)
        with_fluentd="true"
        echo "[WARNING] Fluentd support has been deprecated and will be removed in a future release." >&2
        if ! fluentd_supported; then
          echo "[WARNING] Ignoring the --with-fluentd option since fluentd is currently not supported for ${distro}:${distro_version} ${distro_arch}." >&2
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
      --with-systemd-instrumentation)
        with_systemd_instrumentation="true"
        ;;
      --without-systemd-instrumentation)
        with_systemd_instrumentation="false"
        ;;
      --with-instrumentation-sdk)
        for sdk in $(echo "$2" | tr ',' ' '); do
          if item_in_list "$sdk" "$all_sdks"; then
            with_sdks=$( add_item_to_list "$sdk" "$with_sdks" )
          else
            usage
            echo "[ERROR] Unknown instrumentation SDK: $sdk" >&2
            exit 1
          fi
        done
        shift 1
        ;;
      --without-instrumentation-sdk)
        for sdk in $(echo "$2" | tr ',' ' '); do
          if item_in_list "$sdk" "$all_sdks"; then
            without_sdks=$( add_item_to_list "$sdk" "$without_sdks" )
          else
            usage
            echo "[ERROR] Unknown instrumentation SDK: $sdk" >&2
            exit 1
          fi
        done
        shift 1
        ;;
      --npm-path)
        npm_path="$2"
        if ! command -v "$npm_path" >/dev/null 2>&1; then
          echo "[ERROR] $npm_path not found!" >&2
          exit 1
        fi
        shift 1
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
      --otlp-endpoint)
        otlp_endpoint="$2"
        shift 1
        ;;
      --otlp-endpoint-protocol)
        otlp_endpoint_protocol="$2"
        shift 1
        ;;
      --metrics-exporter)
        metrics_exporter="$2"
        shift 1
        ;;
      --logs-exporter)
        logs_exporter="$2"
        shift 1
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
      --discovery)
        discovery="true"
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
      check_support
      uninstall "$npm_path"
      exit 0
  fi

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

  if [ "$with_fluentd" != "true" ] || ! fluentd_supported; then
    td_agent_version=""
  fi

  check_support

  # check auto instrumentation options
  if [ "$with_instrumentation" = "false" ] && [ "$with_systemd_instrumentation" = "false" ]; then
    instrumentation_version=""
    sdks_to_enable=""
  elif [ "$with_instrumentation" = "true" ] && [ "$with_systemd_instrumentation" = "true" ]; then
    echo "[ERROR] Both --with-instrumentation and --with-systemd-instrumentation options were specified. Only one option is allowed." >&2
    exit 1
  elif ! version_supported "0.87.0" "$instrumentation_version" "$stage"; then
    echo "[ERROR] Unsupported auto instrumentation version: $instrumentation_version" >&2
    echo "[ERROR] Java and Node.js auto instrumentation require version 0.87.0 or greater." >&2
    echo "[ERROR] .NET auto instrumentation requires version 0.99.0 or greater." >&2
    exit 1
  else
    if [ -z "$with_sdks" ] && [ -z "$without_sdks" ]; then
      # user did not explicitly specify any sdk to enable/disable; implicitly enable all supported sdks
      for sdk in $all_sdks; do
        case $sdk in
          java)
            sdks_to_enable=$( add_item_to_list "$sdk" "$sdks_to_enable" )
            ;;
          node)
            if command -v "$npm_path" >/dev/null 2>&1; then
              sdks_to_enable=$( add_item_to_list "$sdk" "$sdks_to_enable" )
            fi
            ;;
          dotnet)
            if dotnet_supported && version_supported "0.99.0" "$instrumentation_version" "$stage"; then
              sdks_to_enable=$( add_item_to_list "$sdk" "$sdks_to_enable" )
            fi
            ;;
        esac
      done
    else
      # fail immediately if any user-specified sdk is not supported
      for sdk in $all_sdks; do
        if item_in_list "$sdk" "$with_sdks" && ! item_in_list "$sdk" "$without_sdks"; then
          case $sdk in
            java)
              sdks_to_enable=$( add_item_to_list "$sdk" "$sdks_to_enable" )
              ;;
            node)
              if ! command -v "$npm_path" >/dev/null 2>&1; then
                echo "[ERROR] npm is required for Node.js auto instrumentation, but was not found." >&2
                echo "[ERROR] Use the '--npm-path <path>' option to specify the absolute path to npm." >&2
                exit 1
              fi
              sdks_to_enable=$( add_item_to_list "$sdk" "$sdks_to_enable" )
              ;;
            dotnet)
              if ! dotnet_supported; then
                echo "[ERROR] .NET auto instrumentation is not currently supported on ${distro_arch}." >&2
                exit 1
              elif ! version_supported "0.99.0" "$instrumentation_version" "$stage"; then
                echo "[ERROR] .NET auto instrumentation requires version 0.99.0 or greater." >&2
                exit 1
              fi
              sdks_to_enable=$( add_item_to_list "$sdk" "$sdks_to_enable" )
              ;;
          esac
        fi
      done
    fi
    if [ -z "$sdks_to_enable" ]; then
      echo "[ERROR] At least one supported SDK must be enabled for auto instrumentation." >&2
      exit 1
    fi
  fi

  ensure_not_installed "$with_fluentd" "$with_instrumentation" "$with_systemd_instrumentation" "$npm_path"

  echo "Splunk OpenTelemetry Collector Version: ${collector_version}"
  echo "Memory Size in MIB: $memory"

  if [ -n "$listen_interface" ]; then
    echo "Listen network interface: $listen_interface"
  fi
  echo "Realm: $realm"
  echo "Ingest Endpoint: $ingest_url"
  echo "API Endpoint: $api_url"
  echo "HEC Endpoint: $hec_url"
  if [ -n "$td_agent_version" ]; then
    echo "TD Agent (Fluentd) Version: $td_agent_version"
  fi
  if [ -n "$sdks_to_enable" ]; then
    echo "Splunk OpenTelemetry Auto Instrumentation Version: $instrumentation_version"
    echo "  Supported Auto Instrumentation SDK(s) to activate: $sdks_to_enable"
    if [ -n "$deployment_environment" ]; then
      echo "  Deployment environment: $deployment_environment"
    else
      echo "  Deployment environment: unknown"
    fi
    if [ -n "$service_name" ]; then
      echo "  Service name: $service_name"
    else
      echo "  Service name: auto-generated"
    fi
    echo "  AlwaysOn Profiling enabled: $enable_profiler"
    echo "  AlwaysOn Memory Profiling enabled: $enable_profiler_memory"
    if [ -n "$otlp_endpoint" ]; then
      echo "  OTLP Endpoint: $otlp_endpoint"
    fi
    if [ -n "$otlp_endpoint_protocol" ]; then
      echo "  OTLP Endpoint Protocol: ${otlp_endpoint_protocol}"
    fi
  fi
  echo

  if [ "${VERIFY_ACCESS_TOKEN:-true}" = "true" ] && ! verify_access_token "$access_token" "$ingest_url" "$insecure"; then
    echo "Your access token could not be verified. This may be due to a network connectivity issue or an invalid access token." >&2
    exit 1
  fi

  install "$stage" "$collector_version" "$td_agent_version" "$skip_collector_repo" "$skip_fluentd_repo" "$instrumentation_version"

  if [ "$with_instrumentation" = "true" ]; then
    if item_in_list "java" "$sdks_to_enable"; then
      create_zeroconfig_java
      sdks_enabled=$( add_item_to_list "java" "$sdks_enabled" )
    elif [ -f "$java_zeroconfig_path" ]; then
      backup_file "$java_zeroconfig_path"
      rm -f "$java_zeroconfig_path"
    fi
    if item_in_list "node" "$sdks_to_enable" && install_node_package "$npm_path"; then
      create_zeroconfig_node
      node_package_installed="true"
      sdks_enabled=$( add_item_to_list "node" "$sdks_enabled" )
    elif [ -f "$node_zeroconfig_path" ]; then
      backup_file "$node_zeroconfig_path"
      rm -f "$node_zeroconfig_path"
    fi
    if item_in_list "dotnet" "$sdks_to_enable" && [ -f "$dotnet_agent_path" ]; then
      create_zeroconfig_dotnet
      sdks_enabled=$( add_item_to_list "dotnet" "$sdks_enabled" )
    elif [ -f "$dotnet_zeroconfig_path" ]; then
      backup_file "$dotnet_zeroconfig_path"
      rm -f "$dotnet_zeroconfig_path"
    fi
    if [ -n "$sdks_enabled" ]; then
      if [ -f "$systemd_instrumentation_config_path" ]; then
        # backup and remove the systemd config if it exists to avoid conflicts with /etc/ld.so.preload
        backup_file "$systemd_instrumentation_config_path"
        rm -f "$systemd_instrumentation_config_path"
      fi
      # add libsplunk.so to /etc/ld.so.preload if it was not added automatically by the instrumentation package
      enable_preload
    fi
  elif [ "$with_systemd_instrumentation" = "true" ]; then
    # remove libsplunk.so from /etc/ld.so.preload if it was added automatically by the instrumentation package
    disable_preload
    sdks_enabled="$sdks_to_enable"
    if item_in_list "node" "$sdks_to_enable" && install_node_package "$npm_path"; then
      node_package_installed="true"
    else
      sdks_enabled=$( remove_item_from_list "node" "$sdks_enabled" )
    fi
    if item_in_list "dotnet" "$sdks_to_enable" && [ ! -f "$dotnet_agent_path" ]; then
      sdks_enabled=$( remove_item_from_list "dotnet" "$sdks_enabled" )
    fi
    if [ -n "$sdks_enabled" ]; then
      create_systemd_instrumentation_config "$sdks_enabled"
    else
      backup_file "$systemd_instrumentation_config_path"
      rm -f "$systemd_instrumentation_config_path"
    fi
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
    echo "[ERROR] The installed splunk-otel-collector package does not include a supported config file!" >&2
    exit 1
  elif [ ! -f "$collector_config_path" ]; then
    echo "[ERROR] Config file $collector_config_path not found!" >&2
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

  if [ -n "$listen_interface" ]; then
    configure_env_file "SPLUNK_LISTEN_INTERFACE" "$listen_interface" "$collector_env_path"
  fi
  configure_env_file "SPLUNK_CONFIG" "$collector_config_path" "$collector_env_path"
  configure_env_file "SPLUNK_ACCESS_TOKEN" "$access_token" "$collector_env_path"
  configure_env_file "SPLUNK_REALM" "$realm" "$collector_env_path"
  configure_env_file "SPLUNK_API_URL" "$api_url" "$collector_env_path"
  configure_env_file "SPLUNK_INGEST_URL" "$ingest_url" "$collector_env_path"
  configure_env_file "SPLUNK_HEC_URL" "$hec_url" "$collector_env_path"
  configure_env_file "SPLUNK_HEC_TOKEN" "$hec_token" "$collector_env_path"
  configure_env_file "SPLUNK_MEMORY_TOTAL_MIB" "$memory" "$collector_env_path"
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

  if [ "$discovery" = "true" ]; then
    configure_env_file "OTELCOL_OPTIONS" "--discovery" "$collector_env_path"
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

  if [ -n "$td_agent_version" ]; then
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

  echo
  cat <<EOH
The Splunk OpenTelemetry Collector for Linux has been successfully installed.

Make sure that your system's time is relatively accurate or else datapoints may not be accepted.

The collector's main configuration file is located at $collector_config_path,
and the environment file is located at $collector_env_path.

If either $collector_config_path or $collector_env_path are modified, the collector service
must be restarted to apply the changes by running the following command as root:

  systemctl restart splunk-otel-collector

EOH

  if [ -n "$td_agent_version" ] && [ -f "$fluent_config_path" ]; then
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

  if [ -n "$sdks_to_enable" ]; then
    if [ -n "$sdks_enabled" ]; then
      if [ "$with_instrumentation" = "true" ]; then
        cat <<EOH
The Splunk OpenTelemetry Auto Instrumentation package has been installed.
/etc/ld.so.preload has been configured for the instrumentation library at $instrumentation_so_path.

The configuration file(s) are located in /etc/splunk/zeroconfig/.

Reboot the system or restart the application(s) for auto instrumentation to take effect.

EOH
      elif [ "$with_systemd_instrumentation" = "true" ]; then
        cat <<EOH
The Splunk OpenTelemetry Auto Instrumentation package has been installed.
Systemd has been configured for auto instrumentation within the
$systemd_instrumentation_config_path drop-in file.

Reboot the system or restart the service(s) for auto instrumentation to take effect.

EOH
      fi
    fi
    if item_in_list "node" "$sdks_to_enable" && ! item_in_list "node" "$sdks_enabled"; then
      if ! command -v "$npm_path" >/dev/null 2>&1; then
        cat <<EOH >&2
[WARNING] Auto Instrumentation for Node.js was not installed since npm was not found.
EOH
      elif [ "$node_package_installed" = "false" ]; then
        cat <<EOH >&2
[WARNING] Auto Instrumentation for Node.js failed installation. Check the output above for details.
EOH
      fi
    fi
    if item_in_list "dotnet" "$sdks_to_enable" && ! item_in_list "dotnet" "$sdks_enabled"; then
      if ! dotnet_supported; then
        cat <<EOH >&2
[WARNING] Auto Instrumentation for .NET was not activated since it is not supported for ${distro_arch}.
EOH
      elif [ ! -f "$dotnet_agent_path" ]; then
        cat <<EOH >&2
[WARNING] Auto Instrumentation for .NET was not activated since it is not supported with the installed instrumentation package version $( get_package_version splunk-otel-auto-instrumentation ).
EOH
      fi
    fi
  fi

  if [ "$with_fluentd" = "true" ] && ! fluentd_supported; then
    cat <<EOH >&2
[WARNING] Fluentd was not installed since it is currently not supported for ${distro}:${distro_version} ${distro_arch}
EOH
  fi

  if [ -z "$listen_interface" ] && [ "$mode" = "agent" ]; then
    echo "[NOTICE] Starting with version 0.86.0, the collector installer changed its default network listening interface from 0.0.0.0 to 127.0.0.1 for agent mode with default configuration. Please consult the release notes for more information and configuration options."
  fi

  exit 0
}

parse_args_and_install "$@"
