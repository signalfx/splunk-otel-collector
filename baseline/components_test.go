// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package baseline

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
)

// TestNewBaselineBuilds verifies the baseline assembles cleanly and is
// non-empty across all component kinds.
func TestNewBaselineBuilds(t *testing.T) {
	factories, err := NewBaseline().Build()
	require.NoError(t, err)

	assert.NotEmpty(t, factories.Receivers)
	assert.NotEmpty(t, factories.Processors)
	assert.NotEmpty(t, factories.Exporters)
	assert.NotEmpty(t, factories.Extensions)
	assert.NotEmpty(t, factories.Connectors)
	assert.NotNil(t, factories.Telemetry)

	assertModuleMetadata(t, factories.Receivers, factories.ReceiverModules)
	assertModuleMetadata(t, factories.Processors, factories.ProcessorModules)
	assertModuleMetadata(t, factories.Exporters, factories.ExporterModules)
	assertModuleMetadata(t, factories.Extensions, factories.ExtensionModules)
	assertModuleMetadata(t, factories.Connectors, factories.ConnectorModules)

	const healthcheckModulePath = "github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	assert.Equal(
		t,
		generatedModuleFallback(t, baselineExtensionModules, healthcheckModulePath),
		factories.ExtensionModules[component.MustNewType("health_check")],
	)
}

// TestBaselineIsUpstreamOnly asserts the baseline carries no Splunk-specific
// components. Those are each flavor's delta, layered on top by the consumer.
func TestBaselineIsUpstreamOnly(t *testing.T) {
	factories, err := NewBaseline().Build()
	require.NoError(t, err)

	for _, splunk := range []string{
		"smartagent",
		"splunk_inputs",
		"discovery",
		"gnmi",
		"lightprometheus",
		"signalfxgatewayprometheusremotewrite",
	} {
		typ := component.MustNewType(splunk)
		assert.NotContains(t, factories.Receivers, typ, splunk)
		assert.NotContains(t, factories.Extensions, typ, splunk)
	}
}

func TestFlavorContributionWithModulePath(t *testing.T) {
	exampleType := component.MustNewType("example_contrib_exporter")
	exampleAlias := component.MustNewType("example_contrib_exporter_alias")
	const (
		exampleModulePath = "example.com/flavor/exporter"
		exampleVersion    = "v1.2.3"
		exampleModuleRef  = exampleModulePath + " " + exampleVersion
	)
	factory := exporter.NewFactory(exampleType, func() component.Config { return &struct{}{} })
	aliasSetter, ok := any(factory).(interface{ SetDeprecatedAlias(component.Type) })
	require.True(t, ok)
	aliasSetter.SetDeprecatedAlias(exampleAlias)

	b := NewBaseline(WithModuleVersion(exampleModulePath, exampleVersion))
	b.AddExportersWithModulePath(exampleModulePath, factory)
	factories, err := b.Build()
	require.NoError(t, err)

	assert.Same(t, factories.Exporters[exampleType], factories.Exporters[exampleAlias])
	assert.Equal(t, exampleModuleRef, factories.ExporterModules[exampleType])
	assert.Equal(t, exampleModuleRef, factories.ExporterModules[exampleAlias])
}

func TestFlavorContributionRequiresModulePath(t *testing.T) {
	exampleType := component.MustNewType("example_contrib_exporter")
	factory := exporter.NewFactory(exampleType, func() component.Config { return &struct{}{} })
	b := NewBaseline()
	b.AddExportersWithModulePath("", factory)

	_, err := b.Build()
	assert.ErrorContains(t, err, "module path is empty")
}

func TestModuleResolverUsesFinalSelectedVersion(t *testing.T) {
	const modulePath = "example.com/component"
	resolver := moduleResolverFromBuildInfo(&debug.BuildInfo{
		Deps: []*debug.Module{{Path: modulePath, Version: "v2.3.4"}},
	})

	moduleRef, err := resolver.resolve(modulePath)
	require.NoError(t, err)
	assert.Equal(t, "example.com/component v2.3.4", moduleRef)
}

func TestModuleResolverAllowsFallbackOnlyForDependencylessTests(t *testing.T) {
	const (
		modulePath = "example.com/component"
		fallback   = "example.com/component v1.2.3"
	)
	tests := []struct {
		name    string
		info    debug.BuildInfo
		want    string
		wantErr bool
	}{
		{
			name: "dependency-less non-main package test",
			info: debug.BuildInfo{Path: "example.com/baseline.test"},
			want: fallback,
		},
		{
			name:    "production binary",
			info:    debug.BuildInfo{Path: "example.com/collector"},
			wantErr: true,
		},
		{
			name: "test binary with unrelated dependency metadata",
			info: debug.BuildInfo{
				Path: "example.com/baseline.test",
				Deps: []*debug.Module{{Path: "example.com/other", Version: "v1.0.0"}},
			},
			wantErr: true,
		},
		{
			name: "main package test uses selected dependency",
			info: debug.BuildInfo{
				Path: "example.com/collector.test",
				Deps: []*debug.Module{{Path: modulePath, Version: "v2.3.4"}},
			},
			want: "example.com/component v2.3.4",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolver := moduleResolverFromBuildInfo(&test.info)
			moduleRef, err := resolver.resolveWithTestFallback(modulePath, fallback)
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, moduleRef)
		})
	}
}

func TestModuleResolverUsesRequestedVersionForLocalReplacement(t *testing.T) {
	const modulePath = "example.com/replaced"
	resolver := moduleResolverFromBuildInfo(&debug.BuildInfo{
		Deps: []*debug.Module{{
			Path:    modulePath,
			Version: "v1.2.3",
			Replace: &debug.Module{Path: "../local"},
		}},
	})

	moduleRef, err := resolver.resolve(modulePath)
	require.NoError(t, err)
	assert.Equal(t, modulePath+" v1.2.3", moduleRef)
}

func TestModuleResolverUsesConfiguredVersionForMainModule(t *testing.T) {
	const modulePath = "example.com/collector"
	resolver := moduleResolverFromBuildInfo(&debug.BuildInfo{
		Main: debug.Module{Path: modulePath, Version: "(devel)"},
	})
	resolver.moduleVersions = map[string]string{modulePath: "v1.2.3"}

	moduleRef, err := resolver.resolve(modulePath)
	require.NoError(t, err)
	assert.Equal(t, modulePath+" v1.2.3", moduleRef)
}

func TestModuleResolverRejectsUnversionedModule(t *testing.T) {
	const modulePath = "example.com/versionless"
	resolver := moduleResolverFromBuildInfo(&debug.BuildInfo{
		Deps: []*debug.Module{{Path: modulePath}},
	})

	_, err := resolver.resolve(modulePath)
	assert.Error(t, err)
}

// TestBuildReportsDuplicates verifies a colliding component type surfaces an
// error from MakeFactoryMap rather than silently overwriting.
func TestBuildReportsDuplicates(t *testing.T) {
	exampleType := component.MustNewType("example_contrib_exporter")
	mk := func() exporter.Factory {
		return exporter.NewFactory(exampleType, func() component.Config { return &struct{}{} })
	}
	b := NewBaseline(WithModuleVersion("example.com/flavor/exporter", "v1.2.3"))
	b.AddExportersWithModulePath("example.com/flavor/exporter", mk(), mk())
	_, err := b.Build()
	assert.Error(t, err)
}

func assertModuleMetadata[T component.Factory](
	t *testing.T,
	factories map[component.Type]T,
	modules map[component.Type]string,
) {
	t.Helper()
	assert.Len(t, modules, len(factories))
	for componentType, factory := range factories {
		module, ok := modules[componentType]
		require.True(t, ok, "missing module for %s", componentType)
		assert.NotEmpty(t, module, "empty module for %s", componentType)
		assert.Equal(t, modules[factory.Type()], module, "alias module differs for %s", componentType)
	}
}

func generatedModuleFallback(t *testing.T, modules []moduleMetadata, modulePath string) string {
	t.Helper()
	for _, module := range modules {
		if module.path == modulePath {
			return module.fallback
		}
	}
	t.Fatalf("missing generated module metadata for %q", modulePath)
	return ""
}
