#!/bin/bash

set -euo pipefail

DOTNET_SDK_HOME=/opt/dotnet-sdk
DOTNET_BIN=${DOTNET_SDK_HOME}/dotnet
DOTNET_APP_HOME=/opt/dotnet

useradd -r -m -U -d $DOTNET_APP_HOME -s /bin/false dotnet

wget -nv https://download.visualstudio.microsoft.com/download/pr/19144d78-6f95-4810-a9f6-3bf86035a244/23f4654fc5352e049b517937f94be839/dotnet-sdk-6.0.421-linux-x64.tar.gz -O dotnet-sdk.tar.gz
mkdir -p $DOTNET_SDK_HOME
tar -xzf dotnet-sdk.tar.gz -C $DOTNET_SDK_HOME
rm -f dotnet-sdk.tar.gz
chmod a+x $DOTNET_BIN

wget -nv https://github.com/docker/docker-dotnet-sample/archive/c09abb8c745f336312049db50fcebaefa5a1764a.tar.gz -O docker-dotnet-sample.tar.gz
tar -xzf docker-dotnet-sample.tar.gz -C /tmp
rm -f docker-dotnet-sample.tar.gz
$DOTNET_BIN publish /tmp/docker-dotnet-sample-c09abb8c745f336312049db50fcebaefa5a1764a/src -a x64 -o $DOTNET_APP_HOME
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
