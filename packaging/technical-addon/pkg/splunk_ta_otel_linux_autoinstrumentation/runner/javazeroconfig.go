// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed java-agent-release.txt
var javaVersion string

//go:embed java-agent-sha256sum.txt
var javaAgent256Sum string

const configTemplate = `JAVA_TOOL_OPTIONS=-javaagent:{{.InstrumentationJarPath}}
OTEL_RESOURCE_ATTRIBUTES={{.ResourceAttributes}}
SPLUNK_PROFILER_ENABLED={{.EnableProfiler}}
SPLUNK_PROFILER_MEMORY_ENABLED={{.EnableProfilerMemory}}
SPLUNK_METRICS_ENABLED={{.EnableMetrics}}
{{- if .ServiceName }}
OTEL_SERVICE_NAME={{ .ServiceName }}
{{- end }}
{{- if .OtlpEndpoint }}
OTEL_EXPORTER_OTLP_ENDPOINT={{.OtlpEndpoint}}
{{- end }}
{{- if .OtlpEndpointProtocol }}
OTEL_EXPORTER_OTLP_PROTOCOL={{.OtlpEndpointProtocol}}
{{- end }}
{{- if .MetricsExporter }}
OTEL_METRICS_EXPORTER={{.MetricsExporter}}
{{- end }}
{{- if .LogsExporter }}
OTEL_LOGS_EXPORTER={{.LogsExporter}}
{{- end }}
`

// TemplateData hughesjj@
// I've gone back and forth on whether to just wrap modInputs and add custom as needed
// As of now, only resourceattributes differ from what's default in modInputs
// I can't just add a transformer, given it would differ from the nodejs case
// Since goland does not support autocompletion in the template, I've decided
// to just duplicate them all for now, to decouple from the customer interface
// as defined in inputs.conf
type TemplateData struct {
	InstrumentationJarPath string
	ResourceAttributes     string
	EnableProfiler         string
	EnableProfilerMemory   string
	EnableMetrics          string
	ServiceName            string
	OtlpEndpoint           string
	OtlpEndpointProtocol   string
	MetricsExporter        string
	LogsExporter           string
}

// CreateZeroConfigJava reimplements create_zeroconfig_java from installer script
func CreateZeroConfigJava(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	if "" == modInputs.SplunkOtelJavaAutoinstrumentationJarPath.Value {
		log.Printf("Not instrumenting java, as %s was not set", modInputs.SplunkOtelJavaAutoinstrumentationJarPath.Name)
		return nil
	}
	if "" == modInputs.AutoinstrumentationPath.Value {
		log.Printf("Not instrumenting java, as %s was not set", modInputs.AutoinstrumentationPath.Name)
		return nil
	}
	resourceAttributes := fmt.Sprintf("splunk.zc.method=splunk-otel-auto-instrumentation-%s", strings.TrimSpace(javaVersion))

	if "" != modInputs.DeploymentEnvironment.Value {
		resourceAttributes = fmt.Sprintf("%s,deployment.environment=%s", resourceAttributes, modInputs.DeploymentEnvironment.Value)
	}
	if "" != modInputs.ResourceAttributes.Value {
		resourceAttributes = fmt.Sprintf("%s,%s", resourceAttributes, modInputs.ResourceAttributes.Value)
	}

	tmpl, err := template.New("JavaZeroConfig").Parse(configTemplate)
	if err != nil {
		log.Fatalf("error generating zeroconfig file at %s from template: %#v", modInputs.JavaZeroconfigPath.Value, err)
		return err
	}
	templateData := TemplateData{
		InstrumentationJarPath: modInputs.SplunkOtelJavaAutoinstrumentationJarPath.Value,
		ResourceAttributes:     resourceAttributes,
		EnableProfiler:         modInputs.ProfilerEnabled.Value,
		EnableProfilerMemory:   modInputs.ProfilerMemoryEnabled.Value,
		EnableMetrics:          modInputs.MetricsEnabled.Value,
		ServiceName:            modInputs.OtelServiceName.Value,
		OtlpEndpoint:           modInputs.OtelExporterOtlpEndpoint.Value,
		OtlpEndpointProtocol:   modInputs.OtelExporterOtlpProtocol.Value,
		MetricsExporter:        modInputs.OtelMetricsExporter.Value,
		LogsExporter:           modInputs.OtelLogsExporter.Value,
	}

	if err = os.MkdirAll(filepath.Dir(modInputs.JavaZeroconfigPath.Value), 0755); err != nil {
		err = fmt.Errorf("error creating java zeroconfig path, could not make parent directories: %v", err)
		log.Println(err)
		return err
	}

	filePath, err := os.Create(modInputs.JavaZeroconfigPath.Value)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	if err = tmpl.Execute(filePath, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w %v", err, templateData)
	}
	log.Printf("Successfully generated java autoinstrumentation config at %s\n", filePath.Name())
	return nil
}

func InstrumentJava(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	if "true" != strings.ToLower(modInputs.JavaZeroconfigEnabled.Value) {
		return nil
	}
	if err := CreateZeroConfigJava(modInputs); err != nil {
		return err
	}
	return nil
}

func RemoveJavaInstrumentation(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	if strings.ToLower(modInputs.Backup.Value) != "false" {
		if err := backupFile(modInputs.JavaZeroconfigPath.Value); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Fatalf("error backing up java auto instrumentation configuration, refusing to remove (specify backup=false in inputs.conf if backup not needed): %v", err)
			return err
		}
	}
	if err := os.Remove(modInputs.JavaZeroconfigPath.Value); !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
