#!/bin/sh

# Copyright Splunk Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

getent passwd splunk-otel-collector >/dev/null || \
    useradd --system --user-group --home-dir /etc/otel/collector --no-create-home --shell $(command -v nologin) splunk-otel-collector

if command -v systemctl >/dev/null 2>&1 && systemctl status splunk-otel-collector.service >/dev/null 2>&1; then
    echo "Stopping splunk-otel-collector service"
    systemctl stop splunk-otel-collector.service
    systemctl disable splunk-otel-collector.service
fi
