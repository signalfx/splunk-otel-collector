# Migration example
This example uses `node_exporter`, but the logic should be applicable for any exporter.
Wait 10 seconds after application startup to begin seeing the actual results of the node_exporter scraping.

### Using the removed prometheus_exec receiver
Run `make run-sample-exec` for a "pre-migration" example.

### Using the promtheus scraping receiver
Run `make run-sample-noexec`

