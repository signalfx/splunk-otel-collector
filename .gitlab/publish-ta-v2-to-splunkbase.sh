#!/usr/bin/env bash

# Publishes TA v2 packages to Splunk Base.
#
# Required environment variables:
#   SPLUNKBASE_PASSWORD - Password for the Splunk Base API user
#   APP_ID              - Splunk Base application ID
#   TA_PACKAGES_PATH    - Path to the directory containing .tgz packages
#   VERSION_TAG         - Version tag for release notes (e.g. v0.148.0)

set -uo pipefail

SPLUNK_VERSIONS="8.0,8.1,8.2,9.0,9.1,9.2,9.3,9.4,10.0,10.1,10.2,10.3,10.4"
AUTH="srv-prod-gdi-otel:${SPLUNKBASE_PASSWORD}"
MAX_WAIT_SECONDS=600
POLL_INTERVAL=10

shopt -s nullglob
packages=("${TA_PACKAGES_PATH}"/*.tgz)
shopt -u nullglob

if [ "${#packages[@]}" -eq 0 ]; then
    echo "No .tgz files found in ${TA_PACKAGES_PATH}"
    exit 1
fi

failed=0

for package in "${packages[@]}"; do
    abs_path=$(realpath "$package")
    file_name=$(basename "$package")

    echo "--- Processing ${file_name} ---"

    # Step 1: Upload package and capture the returned id
    echo "Step 1: Uploading ${file_name}..."
    response=$(curl -u "${AUTH}" \
        --request POST "https://splunkbase.splunk.com/api/v1/app/${APP_ID}/new_release" \
        -F "files[]=@${abs_path}" \
        -F "filename=${file_name}" \
        -F "splunk_versions=${SPLUNK_VERSIONS}" \
        -F "visibility=false" \
        -fSs) || { echo "Upload request failed for ${file_name}"; failed=1; continue; }

    id=$(echo "$response" | jq -r '.id')
    if [ -z "$id" ] || [ "$id" = "null" ]; then
        echo "Failed to get id from response: ${response}"
        failed=1
        continue
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

        response=$(curl -u "${AUTH}" \
            --request GET "https://splunkbase.splunk.com/api/v1/package/${id}/" \
            -s)

        result=$(echo "$response" | jq -r '.result')
        echo "  id ${id}: result=${result} (${elapsed}s elapsed)"

        if [ "$result" = "pass" ]; then
            release_file=$(echo "$response" | jq -r '.message.release_file')
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
                error_details=$(echo "$response" | jq -r '
                    .message.error? // .message.details? // .message? // .error? // .details? // empty
                ' 2>/dev/null)
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
        failed=1
        continue
    fi

    # Step 3: Update release notes
    echo "Step 3: Updating release notes for release_file ${release_file}..."
    curl -u "${AUTH}" \
        --request PUT "https://splunkbase.splunk.com/api/v2/apps/${APP_ID}/releases/${release_file}/" \
        --json "{\"release_notes\": \"Add-On with Splunk OpenTelemetry Collector ${VERSION_TAG}\\n\\n[Release Notes](https://github.com/signalfx/splunk-otel-collector/releases/tag/${VERSION_TAG})\"}" \
        || { echo "Failed to update release notes for ${file_name}"; failed=1; continue; }
    echo ""
    echo "Updated release notes for ${file_name} (release_file: ${release_file})"
    echo "--- Done with ${file_name} ---"
done

if [ "$failed" -ne 0 ]; then
    echo "One or more packages failed to publish."
    exit 1
fi

echo "Done!"
