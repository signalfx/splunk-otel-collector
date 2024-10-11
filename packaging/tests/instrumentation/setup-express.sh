#!/bin/bash

set -euo pipefail

EXPRESS_HOME="/opt/express"
useradd -r -m -U -d $EXPRESS_HOME -s /bin/false express

NVM_HOME="/opt/nvm"
mkdir -p $NVM_HOME
HOME=$NVM_HOME bash -c 'curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.5/install.sh | bash'
NVM_DIR="$NVM_HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"

nvm install --default ${NODE_VERSION:-v16}

npm config --global set user root

NODE_PATH="$( npm root -g )"

mkdir -p /etc/profile.d
echo "export PATH=\$PATH:${NVM_BIN}" >> /etc/profile.d/node.sh
echo "export NODE_PATH=${NODE_PATH}" >> /etc/profile.d/node.sh

npm install --global express

cat <<EOH > ${EXPRESS_HOME}/app.js
var express = require('express');
var app = express();

app.get('/', function (req, res) {
   res.send('Hello World');
})

var server = app.listen(3000, function () {
   var host = server.address().address
   var port = server.address().port

   console.log("Example app listening at http://%s:%s", host, port)
})
EOH

chown express:express ${EXPRESS_HOME}/app.js

mkdir -p /etc/systemd/system
cat <<EOH > /etc/systemd/system/express.service
[Unit]
After=network.target

[Service]
Type=simple
User=express
Group=express
Environment=NODE_PATH=${NODE_PATH}
ExecStart=${NVM_BIN}/node ${EXPRESS_HOME}/app.js
ExecStop=/bin/kill -TERM \$MAINPID
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOH
