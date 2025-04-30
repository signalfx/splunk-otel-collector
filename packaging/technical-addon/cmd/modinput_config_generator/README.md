# Building this tool
To build this tool, run `make ta-build-tools`
To run tests for this tool, `make test-ta-build-tools`

# Creating new addons using this tool
To create a new addon,

1. Determine the name of your addon, ex `Sample_Addon`
1. Make a new directory under `packaging/technical-addon/pkg/` with the lower-cased name of your new addon (ex `packaging/technical-addon/pkg/Sample_Addon`)
1. In this newly created directory, create a new `runners/modular-inputs.yaml` (see below for details)
1. In this newly created directory, create a new `assets/inputs.conf.tmpl` and `assets/inputs.conf.spec.tmpl`.  See below for examples. Edit these as nescessary, but the defaults should be good enough for most cases.
1. In this newly created directory, create a new main module under ex `runner/main.go`.  This runner will actually invoke whatever logic you wish the addon to perform.
1. If you need windows/linux specific code, feel free to use golang build flags to implement any such need.
1. Run `make build-ta-runners`
1. Your addon layout will now live under `$BUILD_DIR/Sample_Addon` (or whatever you named it) (TODO make this live under a `$BUILD_DIR/technical/addons` folder).  You can tar -xzvf this to your hearts content as you normally would to create an addon.

## Example inputs.conf.tmpl
```
[{{ .SchemaName }}://{{ .SchemaName }}]
disabled=false
start_by_shell=false
interval = 60
index = _internal
sourcetype = {{ .SchemaName }}

{{- range $name, $inputConfig := .ModularInputs }}
{{- if $inputConfig.Default }}
{{ $name }} =  "{{ $inputConfig.Default }}"
{{- else if $inputConfig.Required }}
{{ $name }} =
{{- end }}
{{- end }}
```

## Example inputs.conf.spec.tmpl
```
[{{.SchemaName}}://<name>]
{{ range $name, $config := .ModularInputs }}
{{ $name }} = <value>
* {{ $config.Description }} {{ if $config.Default }} (Default: {{$config.Default}} ){{ end }}
{{ end }}
```


## Structure of modular-inputs.yaml
As always, code is source of truth.  Currently, the schema looks like so:

```
modular-input-schema-name: Sample_Addon
modular-inputs:
  everything_set:
    description: "SET ALL THE THINGS"
    default: "$SPLUNK_OTEL_TA_HOME/local/access_token"
    passthrough: true
    replaceable: true
    flag:
      name: "test-flag"
      is-unary: false

  minimal_set:
    description: "This is all you need"

  unary_flag_with_everything_set:
    description: "Unary flags don't take arguments/values and are either present or not"
    default: "$SPLUNK_OTEL_TA_HOME/local/access_token"
    passthrough: true
    replaceable: true
    flag:
      name: "test-flag"
      is-unary: true

  minimal_set_required:
    description: "hello"
    required: true
```


## Example addon golang binary

If you need to invoke another command (such as the collector), you may use golang's exec.Command as such

```go
package main

import (
	"os"
	"os/exec"
)

func main() {

	// modularinput.ModinputProcessor gives GetFlags and GetEnvVars functionality
	var flags []string
	var envVars []string

	cmd := exec.Command("true", flags...)
	cmd.Env = append(os.Environ(), envVars...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
```

Alternatively, if you need platform specific code, you may use `//go:build` flags, providing a file for every needed variation of code.