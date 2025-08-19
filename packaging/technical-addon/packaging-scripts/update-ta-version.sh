#!/bin/bash -eux
APP_CONFIG="$ADDONS_SOURCE_DIR/Splunk_TA_otel/default/app.conf"
if ! [ -f "$APP_CONFIG" ]; then
    echo "file $APP_CONFIG not found"
    exit 1
fi
# for sed command, the capturing group for version is what will be printed (returned), but it'll print every line, so only grab first
OLD_TA_VERSION="$(sed -n 's/^version = \(.*$\)/\1/p' "$APP_CONFIG" | head -n 1)"
if [ -z "$TA_VERSION" ]; then
    TA_VERSION=$(echo "$OLD_TA_VERSION" | awk -F. '{printf "%d.%d.%d", $1,$2+1,0}')
fi
echo "Updating TA from $OLD_TA_VERSION to version $TA_VERSION"
sed -i'.old' "s/^version = .*$/version = $TA_VERSION/g" "$APP_CONFIG" && rm "$APP_CONFIG.old"
