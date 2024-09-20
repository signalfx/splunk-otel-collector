#!/bin/bash -eux

echo "THIS COMMAND SHOULD ONLY BE RUN ON A NON-PUBLICLY-RELEASED TAG VERSION"

APP_CONFIG="$BUILD_DIR/Splunk_TA_otel/default/app.conf"
if ! [ -f "$APP_CONFIG" ]; then
    echo "file $APP_CONFIG not found"
    exit 1
fi

TA_VERSION="$(sed -n 's/^version = \(.*$\)/\1/p' "$APP_CONFIG" | head -n 1)"
[[ -z "$TA_VERSION" ]] && exit 1

echo "DELETING TAG+BRANCH (release/)v$TA_VERSION FROM LOCAL AND ORIGIN (if exists)"
git tag --delete "v$TA_VERSION" || true
git push --delete origin "v$TA_VERSION" || true
git push --delete origin "release/v$TA_VERSION" || true
