---
name: prepare-changelog
description: Prepare the changelog for the next release.
---

Run make to prepare the changelog for the next release:

```bash
make prepare-changelog VERSION<version>
```

Replace `<version>` with the version number for the next release, e.g.: v0.154.0

This will update the CHANGELOG.md file with the new version and the changes since the last release.

Remove irrelevant entries from CHANGELOG.md, exclude entries that:
- Are not listed in `components.md``
- Lack direct reference to the splunk-otel-collector repository
- Are not affecting our customers in any way
- Examples to exclude:
  - updates to `cmd/*`components from upstream, like `cmd/telemetrygen`
  - updates to test utilities, like `testbed`
  - `mdatagen` updates

Collect everything that was excluded into a new file named `release-excluded-changes.md`

Update product names in the lines added to `CHANGELOG.md` that:
- Ensure product names use proper public names (capitalization, spacing, etc.)
- For example, "Rabbitmq" should be updated to "RabbitMQ"
- Do not change casing of names referring to component names in YAML configurations, nor
  names of components in source code.

Ensure consistency in formatting and style of the changelog entries, following the existing format in `CHANGELOG.md`.

Examples of PRs that should be used as reference for the above instructions:
- https://github.com/signalfx/splunk-otel-collector/pull/7590
- https://github.com/signalfx/splunk-otel-collector/pull/7553
- https://github.com/signalfx/splunk-otel-collector/pull/7523
