#!/bin/bash

set -e

# Script to prepare changelog for a new release
# Usage: ./prepare-changelog.sh <version>

if [ $# -eq 0 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.133.0"
    exit 1
fi

VERSION="$1"
TEMP_DIR=$(mktemp -d)
CHANGELOG_FILE="CHANGELOG.md"

# Category headers
HEADER_BREAKING="### 🛑 Breaking changes 🛑"
HEADER_DEPRECATIONS="### 🚩 Deprecations 🚩"
HEADER_NEW_COMPONENTS="### 🚀 New components 🚀"
HEADER_ENHANCEMENTS="### 💡 Enhancements 💡"
HEADER_BUG_FIXES="### 🧰 Bug fixes 🧰"


# Function to check if a number is a PR or issue using redirect behavior
check_pr_or_issue() {
    local repo_url="$1"
    local number="$2"
    
    # Check if /pull/X redirects to /issues/X (meaning it's an issue, not a PR)
    local redirect=$(curl -I -s "$repo_url/pull/$number" 2>/dev/null | grep -i "^location:" | cut -d' ' -f2 | tr -d '\r' || true)
    
    # If it redirects to /issues/, then it's an issue
    if [[ "$redirect" == *"/issues/"* ]]; then
        echo "issues"
        return
    fi
    
    # Otherwise, assume it's a PR
    echo "pull"
}

# Function to convert PR/issue numbers to proper markdown links
convert_pr_issue_links() {
    local content="$1"
    local repo_url="$2"

    # Extract all (#12345) patterns and convert them one by one
    local result="$content"
    
    # Find all (#number) patterns
    local numbers=$(echo "$content" | grep -oE '\(#[0-9]+\)' | grep -oE '[0-9]+' | sort -u)
    
    local total=$(echo "$numbers" | wc -w)
    
    for number in $numbers; do
        local type=$(check_pr_or_issue "$repo_url" "$number")
        
        # Replace (#number) with proper markdown link
        result=$(echo "$result" | sed "s|(#$number)|([#$number]($repo_url/$type/$number))|g")
    done
    
    echo "$result"
}


# Function to fetch and process upstream changelog entries
fetch_upstream_entries() {
    local repo_url="$1"
    local prefix="$2"
    local version="$3"
    
    
    # Download the changelog
    local changelog_url="$repo_url/raw/main/CHANGELOG.md"
    local temp_changelog="$TEMP_DIR/$(basename "$repo_url")_changelog.md"
    
    if ! curl -L -s -o "$temp_changelog" "$changelog_url"; then
        echo "Error: Failed to download changelog from $repo_url" >&2
        return 1
    fi
    
    # Extract entries for the specific version (handle both "## v0.132.0" and "## v1.38.0/v0.132.0" formats)
    local escaped_version="${version//\./\\.}"
    local entries=$(sed -E -n "/^## .*${escaped_version}/,/^(## |<!-- previous-version -->)/p" "$temp_changelog" | sed '$d' | tail -n +3)
    
    if [ -z "$entries" ]; then
        echo "Warning: No entries found for version $version in $repo_url" >&2
        return 0
    fi
    
    # Add prefix to each entry
    local prefixed_entries=""
    while IFS= read -r line; do
        if [[ "$line" =~ ^-[[:space:]] ]]; then
            prefixed_entries="$prefixed_entries- ($prefix) ${line:2}"$'\n'
        elif [[ "$line" =~ ^###[[:space:]] ]]; then
            prefixed_entries="$prefixed_entries$line"$'\n'
        elif [ -n "$line" ]; then
            prefixed_entries="$prefixed_entries$line"$'\n'
        else
            prefixed_entries="$prefixed_entries"$'\n'
        fi
    done <<< "$entries"
    
    # Convert PR/issue references to proper markdown links
    prefixed_entries=$(convert_pr_issue_links "$prefixed_entries" "$repo_url")
    
    echo "$prefixed_entries"
}

# Helper function to extract section content
extract_section() {
    local content="$1"
    local header="$2"
    
    # Use awk to handle special characters in headers properly
    echo "$content" | awk -v header="$header" '
    BEGIN { in_section = 0 }
    $0 == header { in_section = 1; next }
    /^### / && in_section { exit }
    in_section && NF > 0 { print }
    '
}

# Function to parse entries into categories  
parse_entries_by_category() {
    local content="$1"
    local temp_dir=$(mktemp -d)
    
    # Extract each section by header
    extract_section "$content" "$HEADER_BREAKING" > "$temp_dir/breaking" 2>/dev/null || true
    extract_section "$content" "$HEADER_DEPRECATIONS" > "$temp_dir/deprecations" 2>/dev/null || true
    extract_section "$content" "$HEADER_NEW_COMPONENTS" > "$temp_dir/new_components" 2>/dev/null || true
    extract_section "$content" "$HEADER_ENHANCEMENTS" > "$temp_dir/enhancements" 2>/dev/null || true
    extract_section "$content" "$HEADER_BUG_FIXES" > "$temp_dir/bug_fixes" 2>/dev/null || true
    
    
    echo "$temp_dir"
}

# Function to merge entries by category in correct order (Splunk, Core, Contrib)
merge_entries_by_category() {
    local splunk_entries="$1"
    local core_entries="$2" 
    local contrib_entries="$3"
    
    # Parse each set of entries into categories
    local splunk_dir=$(parse_entries_by_category "$splunk_entries")
    local core_dir=$(parse_entries_by_category "$core_entries")
    local contrib_dir=$(parse_entries_by_category "$contrib_entries")
    
    # Standard category order
    local categories=(
        "breaking"
        "deprecations"
        "new_components"
        "enhancements"
        "bug_fixes"
    )
    
    local result=""
    
    # For each category, merge entries in order: Splunk, Core, Contrib
    for category in "${categories[@]}"; do
        local category_content=""
        local has_entries=false
        
        # Add Splunk entries for this category
        if [ -f "$splunk_dir/$category" ]; then
            category_content+="$(cat "$splunk_dir/$category")"
            has_entries=true
        fi
        
        # Add Core entries for this category  
        if [ -f "$core_dir/$category" ]; then
            [ -n "$category_content" ] && category_content+=$'\n'
            category_content+="$(cat "$core_dir/$category")"
            has_entries=true
        fi
        
        # Add Contrib entries for this category
        if [ -f "$contrib_dir/$category" ]; then
            [ -n "$category_content" ] && category_content+=$'\n'
            category_content+="$(cat "$contrib_dir/$category")"
            has_entries=true
        fi
        
        # If we have entries for this category, add the header and content
        if [ "$has_entries" = true ]; then
            [ -n "$result" ] && result+=$'\n'
            case "$category" in
                "breaking")
                    result+="$HEADER_BREAKING"$'\n\n'
                    ;;
                "deprecations")
                    result+="$HEADER_DEPRECATIONS"$'\n\n'
                    ;;
                "new_components")
                    result+="$HEADER_NEW_COMPONENTS"$'\n\n'
                    ;;
                "enhancements")
                    result+="$HEADER_ENHANCEMENTS"$'\n\n'
                    ;;
                "bug_fixes")
                    result+="$HEADER_BUG_FIXES"$'\n\n'
                    ;;
            esac
            result+="$category_content"$'\n'
        fi
    done
    
    # Cleanup temp directories
    rm -rf "$splunk_dir" "$core_dir" "$contrib_dir"
    
    echo "$result"
}

# Function to replace version section in changelog
replace_version_section() {
    local current_changelog="$1"
    local version="$2"
    local replacement_content="$3"
    
    local temp_file=$(mktemp)
    
    # Part 1: Everything before our version
    echo "$current_changelog" | sed "/^## $version/,\$d" > "$temp_file"
    
    # Part 2: Our version section
    echo "## $version" >> "$temp_file"
    echo >> "$temp_file"
    echo "$replacement_content" >> "$temp_file"
    echo >> "$temp_file"
    
    # Part 3: Everything from next version onwards
    echo "$current_changelog" | awk '/^## / && $0 != "## '"$version"'" {found=1} found' >> "$temp_file"
    
    cat "$temp_file"
    rm -f "$temp_file"
}

main() {
    if [ ! -f "$CHANGELOG_FILE" ]; then
        echo "Error: CHANGELOG.md not found" >&2
        exit 1
    fi
    local current_changelog=$(cat "$CHANGELOG_FILE")

    # Take entries added by .chologen, add (Splunk) prefix and make PR/issue links
    local splunk_entries=$(echo "$current_changelog" | sed -E -n "/^## $VERSION/,/^(## |<!-- previous-version -->)/p" | sed '$d' | tail -n +3 | sed 's/^- \([^(]\)/- (Splunk) \1/')
    splunk_entries=$(convert_pr_issue_links "$splunk_entries" "https://github.com/signalfx/splunk-otel-collector")

    local core_entries=$(fetch_upstream_entries "https://github.com/open-telemetry/opentelemetry-collector" "Core" "$VERSION")
    local contrib_entries=$(fetch_upstream_entries "https://github.com/open-telemetry/opentelemetry-collector-contrib" "Contrib" "$VERSION")
    local merged_content=$(merge_entries_by_category "$splunk_entries" "$core_entries" "$contrib_entries")

    # Build the replacement content for the version section
    local replacement_content="This Splunk OpenTelemetry Collector release includes changes from the [opentelemetry-collector $VERSION](https://github.com/open-telemetry/opentelemetry-collector/releases/tag/$VERSION)
and the [opentelemetry-collector-contrib $VERSION](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/$VERSION) releases where appropriate.

$merged_content"

    # Replace the version section with the complete content
    local new_changelog=$(replace_version_section "$current_changelog" "$VERSION" "$replacement_content")
    echo "$new_changelog" > "$CHANGELOG_FILE"

    rm -rf "$TEMP_DIR"
    echo "Changelog preparation completed for version $VERSION"
    echo "Please review $CHANGELOG_FILE and remove unwanted entries before committing"
}

# Ensure cleanup on exit
trap 'rm -rf "$TEMP_DIR"' EXIT

main "$@"
