package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
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
		log.Fatalf("error generating zeroconfig file at %s from template: %#v", modInputs.ZeroconfigPath.Value, err)
		return err
	}
	templateData := TemplateData{
		InstrumentationJarPath: modInputs.SplunkOtelJavaAutoinstrumentationJarPath.Value,
		ResourceAttributes:     resourceAttributes,
		EnableProfiler:         modInputs.ProfilerEnabled.Value,
		EnableProfilerMemory:   modInputs.ProfilerMemoryEnabled.Value,
		EnableMetrics:          modInputs.MetricsEnabled.Value,
		ServiceName:            modInputs.OtelSeviceName.Value,
		OtlpEndpoint:           modInputs.OtelExporterOtlpEndpoint.Value,
		OtlpEndpointProtocol:   modInputs.OtelExporterOtlpProtocol.Value,
		MetricsExporter:        modInputs.OtelMetricsExporter.Value,
		LogsExporter:           modInputs.OtelLogsExporter.Value,
	}
	filePath, err := os.Create(modInputs.ZeroconfigPath.Value)
	if err = tmpl.Execute(filePath, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w %v", err, templateData)
	}

	return nil
}

func InstrumentJava(modInputs *SplunkTAOtelLinuxAutoinstrumentationModularInputs) error {
	if err := CreateZeroConfigJava(modInputs); err != nil {
		return err
	}
	return nil
}
