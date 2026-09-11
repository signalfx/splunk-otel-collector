# Example of deployment with a Splunk instance

This example shows how to use the `splunk_inputs` receiver and `splunk_outputs` exporter together to run a Splunk Technology Add-on (TA) without a real `splunkd`. The receiver reads the TA's `inputs.conf`, `transforms.conf`, and `props.conf` directly; the exporter reads `outputs.conf` to determine where to send the collected logs. Both components layer conf files using standard Splunk conf precedence.

## Directory layout

Both `splunk_inputs` and `splunk_outputs` share the same `base_dir`, which is the root of a standard Splunk Universal Forwarder directory tree:

```
<base_dir>/
  etc/
    apps/
      <TA_name>/
        default/
          inputs.conf
          transforms.conf   # optional
          props.conf        # optional
        local/              # optional — overrides default/
          inputs.conf
    system/
      default/              # built-in system-level defaults
      local/                # optional — site-wide overrides for all TAs
          outputs.conf
```

Conf files are layered in Splunk precedence order: `etc/system/default` < `etc/apps/<TA>/default` < `etc/apps/<TA>/local` < `etc/system/local`.

In this example:
- `Splunk_TA_nix` is mounted at `/var/splunk_home/etc/apps/Splunk_TA_nix`
- `outputs.conf` is mounted at `/var/splunk_home/etc/system/local/outputs.conf`
- `base_dir` for both components is `/var/splunk_home`

Multiple TAs can be mounted under `etc/apps/` at the same time; the receiver discovers and starts them all.

## Deploy a local Splunk instance

In this folder, run:

```
docker compose up -d
```

This deploys a Splunk instance locally available at `http://localhost:18000` with credentials `admin` / `changeme`.

## Download the Splunk Add-on for Unix and Linux

Download the TA from https://splunkbase.splunk.com/app/833 (requires a Splunk account). Save the downloaded `.tgz` file locally.

## Install the TA on the Splunk instance

Go to **Manage Apps** in Splunk Web and install the TA from your downloaded `.tgz` file. This makes Splunk aware of the TA's data models and field extractions so results display correctly.

## Install The Splunk App for Content Packs (optional)

Download and install this app from https://splunkbase.splunk.com/app/5391 to get the dashboards that come with the TA's data.

See https://help.splunk.com/en/splunk-it-service-intelligence/content-packs-for-itsi-and-ite/unix-dashboards-and-reports/1.3/install-the-content-pack-for-unix-dashboards-and-reports for more information.

## Set up the TA for the collector

Extract the downloaded TA into this folder (replace the path as needed):

```
tar xzvf ~/Downloads/splunk-add-on-for-unix-and-linux_1020.tgz
```

This creates a `Splunk_TA_nix` directory. Copy `default/` to `local/` and enable the inputs you want to collect:

```
cd Splunk_TA_nix && cp -r default local
```

Open `local/inputs.conf` and change `disabled = 1` to `disabled = 0` for each stanza you want to enable.

## Build the collector binary

The example Dockerfile copies the collector binary from `bin/otelcol_linux_amd64`. Build it from the repository root first:

```
make otelcol
```

or copy an existing Linux amd64 binary to `bin/otelcol_linux_amd64`.

Both the `splunk_inputs` receiver and `splunk_outputs` exporter require the `enableTARunner` feature gate, which is enabled via `--feature-gates=+enableTARunner` in the collector command (already set in `docker-compose.yml`). Both components are available from `v0.158.0` onward.

## Start the example

From this directory:

```
docker compose up -d
```

## Search the main index

Go to Splunk Web and search `index=main` to see the collected TA data.