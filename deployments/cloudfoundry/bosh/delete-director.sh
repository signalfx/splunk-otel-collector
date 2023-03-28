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

#!/bin/bash

set -euo pipefail

BOSH_DIRECTOR_DIR="./bosh-env/virtualbox"
BOSH_DIRECTOR_DEPLOYMENT_DIR="bosh-deployment"
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -d $BOSH_DIRECTOR_DIR ] && [ -d $BOSH_DIRECTOR_DIR/$BOSH_DIRECTOR_DEPLOYMENT_DIR ]; then
  cd $BOSH_DIRECTOR_DIR
  ./$BOSH_DIRECTOR_DEPLOYMENT_DIR/virtualbox/delete-env.sh
  cd $SCRIPT_DIR
  rm -rf $BOSH_DIRECTOR_DIR
fi