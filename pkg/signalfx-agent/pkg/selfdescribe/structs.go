package selfdescribe

import (
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

// Embedded structs to always ignore since they already provided in other
// places in the output.
var excludedFields = map[string]bool{
	"MonitorConfig":  true,
	"ObserverConfig": true,
	"OtherConfig":    true,
}

// This will have to change if we start pulling in monitors from other repos
func packageDirOfType(t reflect.Type) string {
	return strings.TrimPrefix(t.PkgPath(), "github.com/signalfx/signalfx-agent/")
}

func getStructMetadata(typ reflect.Type) structMetadata {
	packageDir := packageDirOfType(typ)
	structName := typ.Name()
	if packageDir == "" || structName == "" {
		return structMetadata{}
	}

	fieldMD := []fieldMetadata{}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)

		if excludedFields[f.Name] {
			continue
		}

		if f.Anonymous && indirectKind(f.Type) == reflect.Struct {
			nestedSM := getStructMetadata(f.Type)
			fieldMD = append(fieldMD, nestedSM.Fields...)
			continue
			// Embedded struct name and doc is irrelevant.
		}

		yamlName := getYAMLName(f)
		if (yamlName == "" || yamlName == "-" || strings.HasPrefix(yamlName, "_")) && (!isInlinedYAML(f) || indirectKind(f.Type) != reflect.Struct) {
			continue
		}

		fm := fieldMetadata{
			YAMLName: yamlName,
			Doc:      structFieldDocs(packageDir, structName)[f.Name],
			Default:  getDefault(f),
			Required: getRequired(f),
			Type:     indirectKind(f.Type).String(),
		}

		if indirectKind(f.Type) == reflect.Struct {
			smd := getStructMetadata(indirectType(f.Type))
			fm.ElementStruct = &smd
		} else if f.Type.Kind() == reflect.Map || f.Type.Kind() == reflect.Slice {
			if f.Type != reflect.TypeOf(yaml.MapSlice(nil)) {
				ikind := indirectKind(f.Type.Elem())
				fm.ElementKind = ikind.String()

				if ikind == reflect.Struct {
					smd := getStructMetadata(indirectType(f.Type.Elem()))
					fm.ElementStruct = &smd
				}
			}
		}

		fieldMD = append(fieldMD, fm)
	}

	return structMetadata{
		Name:    structName,
		Doc:     structDoc(packageDir, structName),
		Package: packageDir,
		Fields:  fieldMD,
	}
}
