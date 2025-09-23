#!/bin/bash

set -euo pipefail

DOTNET_SDK_HOME=/opt/dotnet-sdk
DOTNET_BIN=${DOTNET_SDK_HOME}/dotnet
DOTNET_APP_HOME=/opt/dotnet

useradd -r -m -U -d $DOTNET_APP_HOME -s /bin/false dotnet

wget -nv https://builds.dotnet.microsoft.com/dotnet/Sdk/8.0.414/dotnet-sdk-8.0.414-linux-x64.tar.gz -O dotnet-sdk.tar.gz
mkdir -p $DOTNET_SDK_HOME
tar -xzf dotnet-sdk.tar.gz -C $DOTNET_SDK_HOME
rm -f dotnet-sdk.tar.gz
chmod a+x $DOTNET_BIN

# The SHA below is the commit targeted for the sample application
wget -nv https://github.com/docker/docker-dotnet-sample/archive/c7f01a5a7f2058bc1e1e29f8cfdb92fd1800054d.tar.gz -O docker-dotnet-sample.tar.gz
tar -xzf docker-dotnet-sample.tar.gz -C /tmp
rm -f docker-dotnet-sample.tar.gz
$DOTNET_BIN publish /tmp/docker-dotnet-sample-c7f01a5a7f2058bc1e1e29f8cfdb92fd1800054d/src -a x64 -o $DOTNET_APP_HOME
chown -R dotnet:dotnet $DOTNET_APP_HOME

mkdir -p /etc/systemd/system
cat <<EOH > /etc/systemd/system/dotnet.service
[Unit]
After=network.target

[Service]
Type=simple
User=dotnet
Group=dotnet
WorkingDirectory=${DOTNET_APP_HOME}
ExecStart=${DOTNET_BIN} ${DOTNET_APP_HOME}/myWebApp.dll
ExecStop=/bin/kill -TERM \$MAINPID
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOH
