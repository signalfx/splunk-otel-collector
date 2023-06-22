//go:build linux
// +build linux

package collectd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"text/template"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/services"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// WriteConfFile writes a file to the given filePath, ensuring that the
// containing directory exists.
func WriteConfFile(content, filePath string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
		return fmt.Errorf("failed to create collectd config dir at %s: %w", filepath.Dir(filePath), err)
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create/truncate collectd config file at %s: %w", filePath, err)
	}
	defer f.Close()

	// Lock the file down since it could contain credentials
	if err := f.Chmod(0600); err != nil {
		return fmt.Errorf("failed to restrict permissions on collectd config file at %s: %w", filePath, err)
	}

	_, err = f.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write collectd config file at %s: %w", filePath, err)
	}

	log.Debugf("Wrote file %s", filePath)

	return nil
}

// InjectTemplateFuncs injects some helper functions into our templates.
func InjectTemplateFuncs(tmpl *template.Template) *template.Template {
	tmpl.Funcs(
		template.FuncMap{
			// Global variables available in all templates
			"Globals": func() map[string]string {
				return map[string]string{
					"Platform": runtime.GOOS,
				}
			},
			"bundleDir": func() string {
				return MainInstance().BundleDir()
			},
			"pluginRoot": func() string {
				return filepath.Join(MainInstance().BundleDir(), "lib/collectd")
			},
			"pythonPluginRoot": func() string {
				return filepath.Join(MainInstance().BundleDir(), "collectd-python")
			},
			"withDefault": func(value interface{}, def interface{}) interface{} {
				v := reflect.ValueOf(value)
				switch v.Kind() {
				case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
					if v.Len() == 0 {
						return def
					}
				case reflect.Ptr:
					if v.IsNil() {
						return def
					}
				case reflect.Invalid:
					return def
				default:
					return value
				}
				return value
			},
			// Makes a slice of the given values
			"sliceOf": func(values ...interface{}) []interface{} {
				return values
			},
			// Encodes dimensions in our "key=value,..." encoding that gets put
			// in the collectd plugin_instance
			"encodeDimsForPluginInstance": func(dims ...map[string]string) (string, error) {
				var encoded []string
				for i := range dims {
					for key, val := range dims[i] {
						encoded = append(encoded, key+"="+val)
					}
				}
				return strings.Join(encoded, ","), nil
			},
			// Encode dimensions for use in an ingest URL
			"encodeDimsAsQueryString": func(dims map[string]string) (string, error) {
				query := url.Values{}
				for k, v := range dims {
					query["sfxdim_"+k] = []string{v}
				}
				return "?" + query.Encode(), nil
			},
			"stringsJoin": strings.Join,
			"stripTrailingSlash": func(s string) string {
				return strings.TrimSuffix(s, "/")
			},
			// Tells whether the key is present in the context map.  Says
			// nothing about whether it is a zero-value or not.
			"hasKey": func(key string, context map[string]interface{}) bool {
				_, ok := context[key]
				return ok
			},
			"merge":           utils.MergeInterfaceMaps,
			"mergeStringMaps": utils.MergeStringMaps,
			"toBool": func(v interface{}) (string, error) {
				if v == nil {
					return strconv.FormatBool(false), nil
				}
				if b, ok := v.(*bool); ok {
					if *b {
						return strconv.FormatBool(true), nil
					}
					return strconv.FormatBool(false), nil
				}
				if b, ok := v.(bool); ok {
					if b {
						return strconv.FormatBool(true), nil
					}
					return strconv.FormatBool(false), nil
				}
				return "", fmt.Errorf("value %#v cannot be converted to bool", v)
			},
			"toMap": utils.ConvertToMapViaYAML,
			"toServiceID": func(s string) services.ID {
				return services.ID(s)
			},
			"toStringMap": utils.InterfaceMapToStringMap,
			"spew": func(obj interface{}) string {
				return spew.Sdump(obj)
			},
			// Renders a subtemplate using the provided context, and optionally
			// a service, which will be added to the context as "service"
			"renderValue": RenderValue,
		})
	return tmpl
}
