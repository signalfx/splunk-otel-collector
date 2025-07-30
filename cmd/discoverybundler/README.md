# Discovery Bundler

The `discoverybundler` command generates discovery configuration files from YAML metadata templates. It creates both the embedded bundle files and the Linux config.d files used by the discovery mode.

## Usage

To generate the latest discovery configuration files:

```bash
make bundle.d
```

This will execute the bundler from `cmd/discoverybundler` and generate:
- Discovery configuration files in `internal/confmapprovider/discovery/bundle.d/`
- Linux config.d files in `cmd/otelcol/config/collector/config.d.linux/`
- Updated embedded filesystem in `internal/confmapprovider/discovery/bundle/generated_bundledfs.go`

## Component Metadata Files

Discovery components are defined using YAML metadata files in the `metadata/` directory:

- `metadata/extensions/` - Extension component definitions  
- `metadata/receivers/` - Receiver component definitions

Each metadata file contains:
- `component_id` - The OpenTelemetry component identifier
- `properties_tmpl` - Template to generate the properties YAML for the component

### Example Metadata File

`metadata/receivers/redis.yaml`:

```yaml
component_id: redis
properties_tmpl: |
  enabled: true
  service_type: redis
  rule:
    docker_observer: type == "container" and port == 6379
    host_observer: type == "hostport" and port == 6379
  config:
    default:
      endpoint: '`endpoint`'
      collection_interval: 10s
  status:
    metrics:
      - status: successful
        strict: redis_connected_clients
        message: Redis receiver is working!
    statements:
      - status: partial
        regexp: 'NOAUTH Authentication required.'
        message: |-
          Make sure your user credentials are correctly specified as an environment variable.
          ```
          {{ configPropertyEnvVar "password" "<password>" }}
          ```
```

## Adding New Components

To add a new discovery component:

1. **Create metadata file**: Add a new YAML file in the appropriate directory:
   - Extensions: `cmd/discoverybundler/metadata/extensions/<component-name>.yaml`
   - Receivers: `cmd/discoverybundler/metadata/receivers/<component-name>.yaml`

2. **Define metadata**: Include the required fields:
   ```yaml
   component_id: <component-name>
   properties_tmpl: |
     enabled: true
     # ... your template content
   ```

3. **Generate files**: Run `make bundle.d` to generate the discovery files

The filename (without extension) automatically becomes the template filename for the generated `.discovery.yaml` file.

## Template Functions

The `properties_tmpl` field supports template functions:

- `{{ configProperty "field" "value" }}` - Creates a property setting
- `{{ configPropertyEnvVar "field" "value" }}` - Creates an environment variable property setting
- `{{ defaultValue }}` - Inserts the default discovery value placeholder

## Generated Output

After running `make bundle.d`, the metadata generates:

1. **Discovery files**: `bundle.d/receivers/redis.discovery.yaml` with templated content:
   ```yaml
   ##############################################################################################
   #                               Do not edit manually!                                        #
   # All changes must be made to associated .yaml metadata file before running 'make bundle.d'. #
   ##############################################################################################
   redis:
     enabled: true
     service_type: redis
     rule:
       docker_observer: type == "container" and port == 6379
     # ... resolved template content with proper env var formatting
   ```

2. **Linux config files**: Commented versions in `cmd/otelcol/config/collector/config.d.linux/`

3. **Embedded filesystem**: `generated_bundledfs.go` file with embedded file lists for all components