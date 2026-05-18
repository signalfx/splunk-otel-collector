#!/usr/bin/env bash

# Publishes a TA v2 package to Splunk Base.
#
# Required environment variables:
#   SPLUNKBASE_PASSWORD - Password for the Splunk Base API user
#   APP_ID              - Splunk Base application ID
#   TA_PACKAGE          - Path to the .tgz package to publish
#   VERSION_TAG         - Version tag for release notes (e.g. v0.148.0)

set -euo pipefail

SPLUNK_VERSIONS="8.0,8.1,8.2,9.0,9.1,9.2,9.3,9.4,10.0,10.1,10.2,10.3,10.4"
AUTH="srv-prod-gdi-otel:${SPLUNKBASE_PASSWORD}"
MAX_WAIT_SECONDS=600
POLL_INTERVAL=10

if [ ! -f "${TA_PACKAGE}" ]; then
    echo "Package file not found: ${TA_PACKAGE}"
    exit 1
fi

if ! abs_path=$(realpath "${TA_PACKAGE}"); then
    echo "Failed to resolve path for ${TA_PACKAGE}"
    exit 1
fi
file_name=$(basename "${TA_PACKAGE}")

echo "--- Processing ${file_name} ---"

# Step 1: Upload package and capture the returned id
echo "Step 1: Uploading ${file_name}..."
response=$(curl -u "${AUTH}" \
    --request POST "https://splunkbase.splunk.com/api/v1/app/${APP_ID}/new_release" \
    -F "files[]=@${abs_path}" \
    -F "filename=${file_name}" \
    -F "splunk_versions=${SPLUNK_VERSIONS}" \
    -F "visibility=false" \
    -fSs) || { echo "Upload request failed for ${file_name}"; exit 1; }

if ! id=$(echo "$response" | jq -r '.id'); then
    echo "Failed to parse upload response for ${file_name}: ${response}"
    exit 1
fi
if [ -z "$id" ] || [ "$id" = "null" ]; then
    echo "Failed to get id from response: ${response}"
    exit 1
fi
echo "Uploaded ${file_name} with id: ${id}"

# Step 2: Poll until result is "pass" or timeout after 10 minutes
echo "Step 2: Waiting for id ${id} to pass validation..."
start_time=$(date +%s)
release_file=""
validation_ok=1

while true; do
    current_time=$(date +%s)
    elapsed=$((current_time - start_time))

    if [ "$elapsed" -ge "$MAX_WAIT_SECONDS" ]; then
        echo "Timeout waiting for id ${id} after ${MAX_WAIT_SECONDS} seconds"
        validation_ok=0
        break
    fi

    if ! response=$(curl -u "${AUTH}" \
        --request GET "https://splunkbase.splunk.com/api/v1/package/${id}/" \
        -fSs); then
        echo "Polling request failed for id ${id}"
        validation_ok=0
        break
    fi

    if ! result=$(echo "$response" | jq -r '.result'); then
        echo "Failed to parse validation response for id ${id}: ${response}"
        validation_ok=0
        break
    fi
    echo "  id ${id}: result=${result} (${elapsed}s elapsed)"

    if [ "$result" = "pass" ]; then
        if ! release_file=$(echo "$response" | jq -r '.message.release_file'); then
            echo "Failed to parse release_file for id ${id}: ${response}"
            validation_ok=0
            break
        fi
        if [ -z "$release_file" ] || [ "$release_file" = "null" ]; then
            echo "Failed to get release_file from response: ${response}"
            validation_ok=0
        else
            echo "  id ${id}: passed, release_file=${release_file}"
        fi
        break
    fi

    case "$result" in
        fail|failed|error)
            if ! error_details=$(echo "$response" | jq -r '
                .message.error? // .message.details? // .message? // .error? // .details? // empty
            ' 2>/dev/null); then
                error_details="$response"
            fi
            if [ -z "$error_details" ] || [ "$error_details" = "null" ]; then
                error_details="$response"
            fi
            echo "Validation failed for id ${id}: ${error_details}"
            validation_ok=0
            break
            ;;
    esac

    sleep "$POLL_INTERVAL"
done

if [ "$validation_ok" -eq 0 ]; then
    exit 1
fi

# Step 3: Update release notes
echo "Step 3: Updating release notes for release_file ${release_file}..."
curl -u "${AUTH}" \
    --request PUT "https://splunkbase.splunk.com/api/v2/apps/${APP_ID}/releases/${release_file}/" \
    --json "{\"release_notes\": \"Add-On with Splunk OpenTelemetry Collector ${VERSION_TAG}\\n\\n[Release Notes](https://github.com/signalfx/splunk-otel-collector/releases/tag/${VERSION_TAG})\"}" \
    -fSs \
    || { echo "Failed to update release notes for ${file_name}"; exit 1; }
echo ""
echo "Updated release notes for ${file_name} (release_file: ${release_file})"
echo "--- Done with ${file_name} ---"
