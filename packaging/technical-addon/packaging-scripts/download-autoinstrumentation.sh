#!/bin/bash -eux
set -o pipefail

[[ -z "$AUTOINSTRUMENTATION_DIR" ]] && echo "AUTOINSTRUMENTATION_DIR not set" && exit 1
[[ -z "$SOURCE_DIR" ]] && echo "SOURCE_DIR not set" && exit 1

JAVA_VERSION="$(cat "${AUTOINSTRUMENTATION_DIR}/packaging/java-agent-release.txt")"
NODEJS_VERSION="$(cat "${AUTOINSTRUMENTATION_DIR}/packaging/nodejs-agent-release.txt")"

if [ "$PLATFORM" == "linux" ] || [ "$PLATFORM" == "all" ]; then
  mkdir -p "${BUILD_DIR}/Splunk_TA_otel_linux_autoinstrumentation/config"
  # Copy our .so, which is made in the build step of the autoinstrumentation Makefile
  cp "${AUTOINSTRUMENTATION_DIR}/dist/libsplunk_amd64.so" "${BUILD_DIR}/Splunk_TA_otel_linux_autoinstrumentation/linux_x86_64/bin/libsplunk_amd64.so"

  java_agent_path="${BUILD_DIR}/Splunk_TA_otel_linux_autoinstrumentation/linux_x86_64/bin/splunk-otel-${NODEJS_VERSION#v}.tgz"
  nodejs_agent_path="${BUILD_DIR}/Splunk_TA_otel_linux_autoinstrumentation/linux_x86_64/bin/splunk-otel-javaagent.jar"
  wget --timestamping "https://github.com/signalfx/splunk-otel-js/releases/download/${NODEJS_VERSION}/splunk-otel-${NODEJS_VERSION#v}.tgz" --output-document "$java_agent_path"
  wget --timestamping "https://github.com/signalfx/splunk-otel-java/releases/download/${JAVA_VERSION}/splunk-otel-javaagent.jar" --output-document  "$nodejs_agent_path"
  # Needed for go:embed
  echo "$JAVA_VERSION" > "${SOURCE_DIR}/pkg/splunk_ta_otel_linux_autoinstrumentation/runner/java-agent-release.txt"
  echo "$NODEJS_VERSION" > "${SOURCE_DIR}/pkg/splunk_ta_otel_linux_autoinstrumentation/runner/nodejs-agent-release.txt"
  sha256sum "$java_agent_path" | cut -d' ' --fields=1 > "${SOURCE_DIR}/pkg/splunk_ta_otel_linux_autoinstrumentation/runner/java-agent-sha256sum.txt"
  sha256sum "$nodejs_agent_path" | cut -d' ' --fields=1  > "${SOURCE_DIR}/pkg/splunk_ta_otel_linux_autoinstrumentation/runner/nodejs-agent-sha256sum.txt"
fi
