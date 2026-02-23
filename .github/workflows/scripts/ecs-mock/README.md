# ECS Metadata Mock

Minimal mock for the ECS Task Metadata V4 endpoint, used by the
`linux-package-test` workflow to smoke-test collector configs that include the
`awsecscontainermetrics` receiver and the `resourcedetection` processor with
the `ecs` detector.

## Files

- `nginx.conf` -- routes `/task` and `/task/stats` to the JSON fixtures.
- `task_metadata.json` -- minimal task metadata response.
- `task_stats.json` -- minimal container stats response.
- `container_metadata.json` -- minimal container metadata response, served at `/`.

## Fixture sources

Simplified from the upstream opentelemetry-collector-contrib test data:

- `internal/aws/ecsutil/ecsutiltest/testdata/task_metadata.json`
- `receiver/awsecscontainermetricsreceiver/testdata/task_stats.json`
