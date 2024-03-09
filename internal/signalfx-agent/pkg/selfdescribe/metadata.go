package selfdescribe

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const monitorMetadataFile = "metadata.yaml"

// MetricMetadata contains a metric's metadata.
type MetricMetadata struct {
	Group       *string `json:"group"`
	Alias       string  `json:"alias,omitempty"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Default     bool    `json:"default" default:"false"`
}

// PropMetadata contains a property's metadata.
type PropMetadata struct {
	Dimension   string `json:"dimension"`
	Description string `json:"description"`
}

// GroupMetadata contains a group's metadata.
type GroupMetadata struct {
	Description string   `json:"description"`
	Metrics     []string `json:"metrics"`
}

// MonitorMetadata contains a monitor's metadata.
type MonitorMetadata struct {
	Dimensions   map[string]DimMetadata    `json:"dimensions"`
	Groups       map[string]*GroupMetadata `json:"groups"`
	Metrics      map[string]MetricMetadata `json:"metrics"`
	Properties   map[string]PropMetadata   `json:"properties"`
	MonitorType  string                    `json:"monitorType" yaml:"monitorType"`
	Doc          string                    `json:"doc"`
	SendAll      bool                      `json:"sendAll" yaml:"sendAll"`
	SendUnknown  bool                      `json:"sendUnknown" yaml:"sendUnknown"`
	NoneIncluded bool                      `json:"noneIncluded" yaml:"noneIncluded"`
}

// PackageMetadata describes a package directory that may have one or more monitors.
type PackageMetadata struct {
	Common      map[string]interface{}
	GoPackage   *string `json:"goPackage" yaml:"goPackage"`
	PackageDir  string  `json:"packageDir" yaml:"packageDir"`
	PackagePath string  `json:"-" yaml:"-"`
	Path        string  `json:"-" yaml:"-"`
	Monitors    []MonitorMetadata
}

// DimMetadata contains a dimension's metadata.
type DimMetadata struct {
	Description string `json:"description"`
}

// CollectMetadata loads metadata for all monitors located in root as well as any subdirectories.
func CollectMetadata(root string) ([]PackageMetadata, error) {
	var packages []PackageMetadata

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || info.Name() != monitorMetadataFile {
			return nil
		}

		var pkg PackageMetadata

		if bytes, err := os.ReadFile(path); err != nil {
			return fmt.Errorf("unable to read metadata file %s: %w", path, err)
		} else if err := yaml.UnmarshalStrict(bytes, &pkg); err != nil {
			return fmt.Errorf("unable to unmarshal file %s: %w", path, err)
		}

		pkg.PackagePath = filepath.Dir(path)
		pkg.Path = path

		packages = append(packages, pkg)

		return nil
	}); err != nil {
		return nil, err
	}

	return packages, nil
}
