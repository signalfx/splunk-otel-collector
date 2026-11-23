#!/bin/bash

set -euo pipefail

mkdir -p dist/assets
cp bin/otelcol_aix_ppc64 dist/assets/
cp bin/otelcol_linux_* dist/assets/
cp bin/otelcol-fips_linux_* dist/assets/
cp instrumentation/dist/libotelinject_*.so dist/assets/
cp dist/signed/* dist/assets/
cp packaging/ta-v2/out/distribution/* dist/assets/
cp packaging/otlpinput/out/distribution/* dist/assets/
(
    cd dist/assets
    rm -f checksums.txt
    sha256sum -- * > checksums.txt
)
