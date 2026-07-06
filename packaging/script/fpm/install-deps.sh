#!/bin/bash

# Copyright Splunk, Inc.
# SPDX-License-Identifier: Apache-2.0

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname ${BASH_SOURCE[0]} )" && pwd )"

apt-get update

apt-get install -y ruby ruby-dev rubygems build-essential git rpm sudo curl jq ruby-bundler python3 python3.13-venv libpython3.13

bundle install --gemfile ${SCRIPT_DIR}/Gemfile

gem install fpm
