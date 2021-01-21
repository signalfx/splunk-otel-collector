# Copyright 2020 Splunk, Inc.
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

from pathlib import Path

# Artifactory
ARTIFACTORY_URL = "https://splunk.jfrog.io/artifactory"
ARTIFACTORY_API_URL = f"{ARTIFACTORY_URL}/api"
ARTIFACTORY_DEB_REPO = "otel-collector-deb"
ARTIFACTORY_DEB_REPO_URL = f"{ARTIFACTORY_URL}/{ARTIFACTORY_DEB_REPO}"
ARTIFACTORY_RPM_REPO = "otel-collector-rpm"
ARTIFACTORY_RPM_REPO_URL = f"{ARTIFACTORY_URL}/{ARTIFACTORY_RPM_REPO}"
DEFAULT_ARTIFACTORY_USERNAME = "otel-collector"

# Signing
CHAPERONE_API_URL = "https://chaperone.re.splunkdev.com/api-service"
DEFAULT_STAGING_USERNAME = "srv-otel-collector"
DEFAULT_TIMEOUT = 600
SIGN_TYPES = ("GPG", "RPM", "WIN")
SIGNED_ARTIFACTS_REPO_URL = "https://repo.splunk.com/artifactory/signed-artifacts"
STAGING_URL = "https://repo.splunk.com/artifactory"
STAGING_REPO = "otel-collector-local"
STAGING_REPO_URL = f"{STAGING_URL}/{STAGING_REPO}"

# Package/Release
REPO_DIR = Path(__file__).parent.parent.parent.parent.parent.parent.resolve()
ASSETS_BASE_DIR = REPO_DIR / "dist" / "release"
COLLECTOR_REPO = "signalfx/splunk-otel-collector"
COMPONENTS = ["deb", "rpm", "windows"]
PACKAGE_NAME = "splunk-otel-collector"
STAGES = ("release", "beta", "test", "github")
S3_BUCKET = "public-downloads--signalfuse-com"
S3_MSI_BASE_DIR = f"{PACKAGE_NAME}/msi"
CLOUDFRONT_DISTRIBUTION_ID = "EJH671JAOI5SN"

# MSI
WIX_IMAGE = "felfert/wix:latest"
WXS_PATH = "internal/buildscripts/packaging/msi/splunk-otel-collector.wxs"
MSI_CONFIG = "cmd/otelcol/config/collector/agent_config_windows.yaml"
