// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configsource

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/knadh/koanf/maps"
	"github.com/spf13/cast"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	// expandPrefixChar is the char used to prefix strings that can be expanded,
	// either environment variables or config sources.
	expandPrefixChar = '$'
	// configSourceNameDelimChar is the char used to terminate the name of config source
	// when it is used to retrieve values to inject in the configuration
	configSourceNameDelimChar = ':'
	// typeAndNameSeparator is the separator that is used between type and name in type/name
	// composite keys.
	typeAndNameSeparator = '/'
	// dollarDollarCompatEnvVar is a temporary env var to disable backward compatibility (true by default)
	dollarDollarCompatEnvVar = "SPLUNK_DOUBLE_DOLLAR_CONFIG_SOURCE_COMPATIBLE"
)

// private error types to help with testability
type (
	errUnknownConfigSource struct{ error }
)

var ddBackwardCompatible = func() bool {
	if v, err := strconv.ParseBool(strings.ToLower(os.Getenv(dollarDollarCompatEnvVar))); err == nil {
		return v
	}
	return true
}()

type ConfigSource interface {
	// Retrieve goes to the configuration source and retrieves the selected data which
	// contains the value to be injected in the configuration and the corresponding watcher that
	// will be used to monitor for updates of the retrieved value. The retrieved value is selected
	// according to the selector and the params arguments.
	//
	// The selector is a string that is required on all invocations, the params are optional. Each
	// implementation handles the generic params according to their requirements.
	Retrieve(ctx context.Context, selector string, params *confmap.Conf, watcher confmap.WatcherFunc) (*confmap.Retrieved, error)
}

// Factory is a factory interface for configuration sources.  Given it's not an accepted component and
// because of the direct Factory usage restriction from https://github.com/open-telemetry/opentelemetry-collector/commit/9631ceabb7dc4ca5cc187bab26d8319783bcc562
// it's not a proper Collector config.Factory.
type Factory interface {
	// CreateDefaultConfig creates the default configuration settings for the ConfigSource.
	// This method can be called multiple times depending on the pipeline
	// configuration and should not cause side-effects that prevent the creation
	// of multiple instances of the ConfigSource.
	// The object returned by this method needs to pass the checks implemented by
	// 'configcheck.ValidateConfig'. It is recommended to have such check in the
	// tests of any implementation of the Factory interface.
	CreateDefaultConfig() Settings

	// CreateConfigSource creates a configuration source based on the given config.
	CreateConfigSource(context.Context, Settings, *zap.Logger) (ConfigSource, error)

	// Type gets the type of the component created by this factory.
	Type() component.Type
}

// Factories maps the type of a ConfigSource to the respective factory object.
type Factories map[component.Type]Factory

// BuildConfigSourcesFromConf inspects the given confmap.Conf and builds all config sources referenced
// in the configuration intended to be used with ResolveWithConfigSources().
func BuildConfigSourcesFromConf(ctx context.Context, confToFurtherResolve *confmap.Conf, logger *zap.Logger, factories Factories, confmapProviders map[string]confmap.Provider) (map[string]ConfigSource, *confmap.Conf, error) {
	configSourceSettings, confWithoutSettings, err := SettingsFromConf(ctx, confToFurtherResolve, factories, confmapProviders)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse settings from conf: %w", err)
	}

	configSources, err := BuildConfigSources(context.Background(), configSourceSettings, logger, factories)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build config sources: %w", err)
	}
	return configSources, confWithoutSettings, nil
}

func BuildConfigSources(ctx context.Context, configSourcesSettings map[string]Settings, logger *zap.Logger, factories Factories) (map[string]ConfigSource, error) {
	cfgSources := make(map[string]ConfigSource, len(configSourcesSettings))
	for fullName, cfgSrcSettings := range configSourcesSettings {
		// If we have the setting we also have the factory.
		factory, ok := factories[cfgSrcSettings.ID().Type()]
		if !ok {
			return nil, fmt.Errorf("unknown %s config source type for %s", cfgSrcSettings.ID().Type(), fullName)
		}

		cfgSrc, err := factory.CreateConfigSource(ctx, cfgSrcSettings, logger.With(zap.String("config_source", fullName)))
		if err != nil {
			return nil, fmt.Errorf("failed to create config source %s: %w", fullName, err)
		}

		if cfgSrc == nil {
			return nil, fmt.Errorf("factory for %q produced a nil extension", fullName)
		}

		cfgSources[fullName] = cfgSrc
	}

	return cfgSources, nil
}

// ResolveWithConfigSources returns a confmap.Conf in which all env vars and config sources on
// the given input config map are resolved to actual literal values of the env vars or config sources.
//
// 1. ResolveWithConfigSources to inject the data from config sources into a configuration;
// 2. Wait for an update on "watcher" func.
// 3. Close the confmap.Retrieved instance;
//
// The current syntax to reference a config source in a YAML is provisional. Currently
// single-line:
//
//	param_to_be_retrieved: $<cfgSrcName>:<selector>[?<params_url_query_format>]
//
// bracketed single-line:
//
//	param_to_be_retrieved: ${<cfgSrcName>:<selector>[?<params_url_query_format>]}
//
// and multi-line are supported:
//
//	param_to_be_retrieved: |
//	  $<cfgSrcName>: <selector>
//	  [<params_multi_line_YAML>]
//
// The <cfgSrcName> is a name string used to identify the config source instance to be used
// to retrieve the value.
//
// The <selector> is the mandatory parameter required when retrieving data from a config source.
//
// Not all config sources need the optional parameters, they are used to provide extra control when
// retrieving and preparing the data to be injected into the configuration.
//
// For single-line format <params_url_query_format> uses the same syntax as URL query parameters.
// Hypothetical example in a YAML file:
//
// component:
//
//	config_field: $file:/etc/secret.bin?binary=true
//
// For multi-line format <params_multi_line_YAML> uses syntax as a YAML inside YAML. Possible usage
// example in a YAML file:
//
// component:
//
//	config_field: |
//	  $yamltemplate: /etc/log_template.yaml
//	  logs_path: /var/logs/
//	  timeout: 10s
//
// Not all config sources need these optional parameters, they are used to provide extra control when
// retrieving and data to be injected into the configuration.
//
// Assuming a config source named "env" that retrieve environment variables and one named "file" that
// retrieves contents from individual files, here are some examples:
//
//	component:
//	  # Retrieves the value of the environment variable LOGS_DIR.
//	  logs_dir: $env:LOGS_DIR
//
//	  # Retrieves the value from the file /etc/secret.bin and injects its contents as a []byte.
//	  bytes_from_file: $file:/etc/secret.bin?binary=true
//
//	  # Retrieves the value from the file /etc/text.txt and injects its contents as a string.
//	  # Hypothetically the "file" config source by default tries to inject the file contents
//	  # as a string if params doesn't specify that "binary" is true.
//	  text_from_file: $file:/etc/text.txt
//
// Bracketed single-line should be used when concatenating a suffix to the value retrieved by
// the config source. Example:
//
//	component:
//	  # Retrieves the value of the environment variable LOGS_DIR and appends /component.log to it.
//	  log_file_fullname: ${env:LOGS_DIR}/component.log
//
// Environment variables are expanded before passed to the config source when used in the selector or
// the optional parameters. Example:
//
//	component:
//	  # Retrieves the value from the file text.txt located on the path specified by the environment
//	  # variable DATA_PATH. The name of the environment variable is the string after the delimiter
//	  # until the first character different than '_' and non-alpha-numeric.
//	  text_from_file: $file:$DATA_PATH/text.txt
//
// Since environment variables and config sources both use the '$', with or without brackets, as a prefix
// for their expansion it is necessary to have a way to distinguish between them. For the non-bracketed
// syntax the code will peek at the first character other than alpha-numeric and '_' after the '$'. If
// that character is a ':' it will treat it as a config source and as environment variable otherwise.
// For example:
//
//	component:
//	  field_0: $PATH:/etc/logs # Injects the data from a config sourced named "PATH" using the selector "/etc/logs".
//	  field_1: $PATH/etc/logs  # Expands the environment variable "PATH" and adds the suffix "/etc/logs" to it.
//
// So if you need to include an environment followed by ':' the bracketed syntax must be used instead:
//
//	component:
//	  field_0: ${PATH}:/etc/logs # Expands the environment variable "PATH" and adds the suffix ":/etc/logs" to it.
//
// For the bracketed syntax the presence of ':' inside the brackets indicates that code will treat the bracketed
// contents as a config source. For example:
//
//	component:
//	  field_0: ${file:/var/secret.txt} # Injects the data from a config sourced named "file" using the selector "/var/secret.txt".
//	  field_1: ${file}:/var/secret.txt # Expands the environment variable "file" and adds the suffix ":/var/secret.txt" to it.
//
// If the character following the '$' is in the set {'*', '#', '$', '@', '!', '?', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
// the code will consider it to be the name of an environment variable to expand, or config source if followed by ':'. Do not use any of these
// characters as the first char on the name of a config source or an environment variable (even if allowed by the system) to avoid unexpected
// results.
//
// For an overview about the internals of the Manager refer to the package README.md.
func ResolveWithConfigSources(ctx context.Context, configSources map[string]ConfigSource, confmapProviders map[string]confmap.Provider, conf *confmap.Conf, watcher confmap.WatcherFunc) (*confmap.Conf, confmap.CloseFunc, error) {
	resolved := map[string]any{}
	var closeFuncs []confmap.CloseFunc
	for _, k := range conf.AllKeys() {
		v := conf.Get(k)
		value, closeFunc, err := resolveConfigValue(ctx, configSources, confmapProviders, v, watcher)
		if err != nil {
			return nil, nil, err
		}
		resolved[k] = value
		if closeFunc != nil {
			closeFuncs = append(closeFuncs, closeFunc)
		}
	}

	maps.IntfaceKeysToStrings(resolved)
	return confmap.NewFromStringMap(resolved), MergeCloseFuncs(closeFuncs), nil
}

// resolveConfigValue takes the value of a "config node" and process it recursively. The processing consists
// in transforming invocations of config sources and/or environment variables into literal data that can be
// used directly from a `confmap.Conf` object.
func resolveConfigValue(ctx context.Context, configSources map[string]ConfigSource, confmapProviders map[string]confmap.Provider, valueToResolve any, watcher confmap.WatcherFunc) (any, confmap.CloseFunc, error) {
	switch v := valueToResolve.(type) {
	case string:
		// Only if the valueToResolve of the node is a string it can contain an env var or config source
		// invocation that requires transformation.
		return resolveStringValue(ctx, configSources, confmapProviders, v, watcher)
	case []any:
		// The valueToResolve is of type []any when an array is used in the configuration, YAML example:
		//
		//  array0:
		//    - elem0
		//    - elem1
		//  array1:
		//    - entry:
		//        str: elem0
		//	  - entry:
		//        str: $tstcfgsrc:elem1
		//
		// Both "array0" and "array1" are going to be leaf config nodes hitting this case.
		nslice := make([]any, 0, len(v))
		var closeFuncs []confmap.CloseFunc
		for _, vint := range v {
			value, closeFunc, err := resolveConfigValue(ctx, configSources, confmapProviders, vint, watcher)
			if err != nil {
				return nil, nil, err
			}
			if closeFunc != nil {
				closeFuncs = append(closeFuncs, closeFunc)
			}
			nslice = append(nslice, value)
		}
		return nslice, MergeCloseFuncs(closeFuncs), nil
	case map[string]any:
		// The valueToResolve is of type map[string]any when an array in the configuration is populated with map
		// elements. From the case above (for type []any) each element of "array1" is going to hit
		// the current case block.
		nmap := make(map[any]any, len(v))
		var closeFuncs []confmap.CloseFunc
		for k, vint := range v {
			value, closeFunc, err := resolveConfigValue(ctx, configSources, confmapProviders, vint, watcher)
			if err != nil {
				return nil, nil, err
			}
			if closeFunc != nil {
				closeFuncs = append(closeFuncs, closeFunc)
			}
			nmap[k] = value
		}
		return nmap, MergeCloseFuncs(closeFuncs), nil
	default:
		// All other literals (int, boolean, etc) can't be further expanded so just return them as they are.
		return v, nil, nil
	}
}

// resolveStringValue transforms environment variables and config sources, if any are present, on
// the given string in the configuration into an object to be inserted into the resulting configuration.
func resolveStringValue(ctx context.Context, configSources map[string]ConfigSource, confmapProviders map[string]confmap.Provider, s string, watcher confmap.WatcherFunc) (any, confmap.CloseFunc, error) {
	var closeFuncs []confmap.CloseFunc

	// Code based on os.Expand function. All delimiters that are checked against are
	// ASCII so bytes are fine for this operation.
	var buf []byte

	// Using i, j, and w variables to keep correspondence with os.Expand code.
	// i tracks the index in s from which a slice to be appended to buf should start.
	// j tracks the char being currently checked and also the end of the slice to be appended to buf.
	// w tracks the number of characters being consumed after a prefix identifying env vars or config sources.
	i := 0
	for j := 0; j < len(s); j++ {
		// Skip chars until a candidate for expansion is found.
		if s[j] == expandPrefixChar && j+1 < len(s) {
			if buf == nil {
				// Assuming that the length of the string will double after expansion of env vars and config sources.
				buf = make([]byte, 0, 2*len(s))
			}

			// Append everything consumed up to the prefix char (but not including the prefix char) to the result.
			buf = append(buf, s[i:j]...)

			var expandableContent, cfgSrcName string
			w := 0 // number of bytes consumed on this pass

			switch {
			case s[j+1] == expandPrefixChar:
				// temporary backward compatibility to support updated $${config_source:value} functionality
				// in provided configs from 0.37.0 until 0.42.0
				bwCompatibilityRequired := false

				var expanded, sourceName string
				var ww int
				if ddBackwardCompatible && len(s[j+1:]) > 2 {
					if s[j+2] == '{' {
						if expanded, ww, sourceName = getBracketedExpandableContent(s, j+2); sourceName != "" {
							bwCompatibilityRequired = true
						}
					} else {
						if expanded, ww, sourceName = getBareExpandableContent(s, j+2); sourceName != "" {
							if len(expanded) > (len(sourceName) + 1) {
								if !strings.Contains(expanded[len(sourceName)+1:], "$") {
									bwCompatibilityRequired = true
								}
							}
						}
					}
				}

				if bwCompatibilityRequired {
					log.Printf(
						`Deprecated config source directive %q has been replaced with %q. Please update your config as necessary as this will be removed in future release. To disable this replacement set the SPLUNK_DOUBLE_DOLLAR_CONFIG_SOURCE_COMPATIBLE environment variable to "false" before restarting the Collector.`,
						s[j:j+2+ww], s[j+1:j+2+ww],
					)
					expandableContent = expanded
					w = ww + 1
					cfgSrcName = sourceName
				} else {
					// Escaping the prefix so $$ becomes a single $ without attempting
					// to treat the string after it as a config source or env var.
					expandableContent = string(expandPrefixChar)
					w = 1 // consumed a single char
				}

			case s[j+1] == '{':
				expandableContent, w, cfgSrcName = getBracketedExpandableContent(s, j+1)

			default:
				expandableContent, w, cfgSrcName = getBareExpandableContent(s, j+1)

			}

			// At this point expandableContent contains a string to be expanded, evaluate and expand it.
			switch {
			case cfgSrcName == "":
				// Not a config source, expand as os.ExpandEnv
				buf = osExpandEnv(buf, expandableContent, w)

			default:
				// A config source, retrieve and apply results.
				retrieved, closeFunc, err := retrieveConfigSourceData(ctx, configSources, confmapProviders, cfgSrcName, expandableContent, watcher)
				if err != nil {
					return nil, nil, err
				}
				if closeFunc != nil {
					closeFuncs = append(closeFuncs, closeFunc)
				}

				consumedAll := j+w+1 == len(s)
				if consumedAll && len(buf) == 0 {
					// This is the only expandableContent on the string, config
					// source is free to return any but parse it as YAML
					// if it is a string or byte slice.
					switch value := retrieved.(type) {
					case []byte:
						if err := yaml.Unmarshal(value, &retrieved); err != nil {
							// The byte slice is an invalid YAML keep the original.
							retrieved = value
						}
					case string:
						if err := yaml.Unmarshal([]byte(value), &retrieved); err != nil {
							// The string is an invalid YAML keep it as the original.
							retrieved = value
						}
					}

					if mapIFace, ok := retrieved.(map[any]any); ok {
						// yaml.Unmarshal returns map[any]any but config
						// map uses map[string]any, fix it with a cast.
						retrieved = cast.ToStringMap(mapIFace)
					}

					return retrieved, MergeCloseFuncs(closeFuncs), nil
				}

				// Either there was a prefix already or there are still characters to be processed.
				if retrieved == nil {
					// Since this is going to be concatenated to a string use "" instead of nil,
					// otherwise the string will end up with "<nil>".
					retrieved = ""
				}

				buf = append(buf, fmt.Sprintf("%v", retrieved)...)
			}

			j += w    // move the index of the char being checked (j) by the number of characters consumed (w) on this iteration.
			i = j + 1 // update start index (i) of next slice of bytes to be copied.
		}
	}

	if buf == nil {
		// No changes to original string, just return it.
		return s, MergeCloseFuncs(closeFuncs), nil
	}

	// Return whatever was accumulated on the buffer plus the remaining of the original string.
	return string(buf) + s[i:], MergeCloseFuncs(closeFuncs), nil
}

func getBracketedExpandableContent(s string, i int) (expandableContent string, consumed int, cfgSrcName string) {
	// Bracketed usage, consume everything until first '}' exactly as os.Expand.
	expandableContent, consumed = scanToClosingBracket(s[i:])
	expandableContent = strings.Trim(expandableContent, " ") // Allow for some spaces.
	delimIndex := strings.Index(expandableContent, string(configSourceNameDelimChar))
	if len(expandableContent) > 1 && delimIndex > -1 {
		// Bracket expandableContent contains ':' treating it as a config source.
		cfgSrcName = expandableContent[:delimIndex]
	}
	return
}

func getBareExpandableContent(s string, i int) (expandableContent string, consumed int, cfgSrcName string) {
	// Non-bracketed usage, ie.: found the prefix char, it can be either a config
	// source or an environment variable.
	var name string
	name, consumed = getTokenName(s[i:])
	expandableContent = name // Assume for now that it is an env var.

	// Peek next char after name, if it is a config source name delimiter treat the remaining of the
	// string as a config source.
	possibleDelimCharIndex := i + consumed
	if possibleDelimCharIndex < len(s) && s[possibleDelimCharIndex] == configSourceNameDelimChar {
		// This is a config source, since it is not delimited it will consume until end of the string.
		cfgSrcName = name
		expandableContent = s[i:]
		consumed = len(expandableContent) // Set consumed bytes to the length of expandableContent
	}
	return
}

// retrieveConfigSourceData retrieves data from the specified config source and injects them into
// the configuration. The Manager tracks sessions and watcher objects as needed.
func retrieveConfigSourceData(ctx context.Context, configSources map[string]ConfigSource, confmapProviders map[string]confmap.Provider, cfgSrcName, cfgSrcInvocation string, watcher confmap.WatcherFunc) (any, confmap.CloseFunc, error) {
	var closeFuncs []confmap.CloseFunc
	cfgSrc, ok := configSources[cfgSrcName]
	var provider confmap.Provider
	var providerFound bool
	if !ok {
		if provider, providerFound = confmapProviders[cfgSrcName]; !providerFound {
			return nil, nil, newErrUnknownConfigSource(cfgSrcName)
		}
	}

	cfgSrcName, selector, paramsConfigMap, err := parseCfgSrcInvocation(cfgSrcInvocation)
	if err != nil {
		return nil, nil, err
	}

	if providerFound {
		if provider == nil {
			return nil, nil, fmt.Errorf("resolving confmap.Provider %q with config sources failed", cfgSrcName)
		}
		retrieved, e := provider.Retrieve(ctx, fmt.Sprintf("%s:%s", cfgSrcName, selector), watcher)
		if e != nil {
			return nil, nil, fmt.Errorf("retrieve error from confmap provider %q: %w", cfgSrcName, e)
		}
		raw, e := retrieved.AsRaw()
		if e != nil {
			return nil, nil, e
		}
		return raw, retrieved.Close, nil
	}

	// Recursively expand the selector.
	expandedSelector, closeFunc, err := resolveStringValue(ctx, configSources, confmapProviders, selector, watcher)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to process selector for config source %q selector %q: %w", cfgSrcName, selector, err)
	}
	if selector, ok = expandedSelector.(string); !ok {
		return nil, nil, fmt.Errorf("processed selector must be a string instead got a %T %v", expandedSelector, expandedSelector)
	}
	if closeFunc != nil {
		closeFuncs = append(closeFuncs, closeFunc)
	}

	// Recursively ResolveWithConfigSources/parse any config source on the parameters.
	if paramsConfigMap != nil {
		paramsConfigMapRet, closeFunc, errResolve := ResolveWithConfigSources(ctx, configSources, confmapProviders, paramsConfigMap, watcher)
		if errResolve != nil {
			return nil, nil, fmt.Errorf("failed to process parameters for config source %q invocation %q: %w", cfgSrcName, cfgSrcInvocation, errResolve)
		}
		if closeFunc != nil {
			closeFuncs = append(closeFuncs, closeFunc)
		}
		paramsConfigMap = confmap.NewFromStringMap(paramsConfigMapRet.ToStringMap())
	}

	retrieved, err := cfgSrc.Retrieve(ctx, selector, paramsConfigMap, watcher)
	if err != nil {
		return nil, nil, fmt.Errorf("config source %q failed to retrieve value: %w", cfgSrcName, err)
	}

	closeFuncs = append(closeFuncs, retrieved.Close)
	val, err := retrieved.AsRaw()
	return val, MergeCloseFuncs(closeFuncs), err
}

func newErrUnknownConfigSource(cfgSrcName string) error {
	return &errUnknownConfigSource{
		fmt.Errorf("config source %q not found", cfgSrcName),
	}
}

// parseCfgSrcInvocation parses the original string in the configuration that has a config source
// retrieve operation and return its "logical components": the config source name, the selector, and
// a confmap.Conf to be used in this invocation of the config source. See Test_parseCfgSrcInvocation
// for some examples of input and output.
// The caller should check for error explicitly since it is possible for the
// other values to have been partially set.
func parseCfgSrcInvocation(s string) (cfgSrcName, selector string, paramsConfigMap *confmap.Conf, err error) {
	parts := strings.SplitN(s, string(configSourceNameDelimChar), 2)
	if len(parts) != 2 {
		err = fmt.Errorf("invalid config source syntax at %q, it must have at least the config source name and a selector", s)
		return
	}
	cfgSrcName = strings.Trim(parts[0], " ")

	// Separate multi-line and single line case.
	afterCfgSrcName := parts[1]
	switch {
	case strings.Contains(afterCfgSrcName, "\n"):
		// Multi-line, until the first \n it is the selector, everything after as YAML.
		parts = strings.SplitN(afterCfgSrcName, "\n", 2)
		selector = strings.Trim(parts[0], " ")

		if len(parts) > 1 && len(parts[1]) > 0 {
			var data map[string]any
			if err = yaml.Unmarshal([]byte(parts[1]), &data); err != nil {
				return
			}
			paramsConfigMap = confmap.NewFromStringMap(data)
		}

	default:
		// Single line, and parameters as URL query.
		const selectorDelim string = "?"
		parts = strings.SplitN(parts[1], selectorDelim, 2)
		selector = strings.Trim(parts[0], " ")

		if len(parts) == 2 {
			paramsPart := parts[1]
			paramsConfigMap, err = parseParamsAsURLQuery(paramsPart)
			if err != nil {
				err = fmt.Errorf("invalid parameters syntax at %q: %w", s, err)
				return
			}
		}
	}

	return cfgSrcName, selector, paramsConfigMap, err
}

func parseParamsAsURLQuery(s string) (*confmap.Conf, error) {
	values, err := url.ParseQuery(s)
	if err != nil {
		return nil, err
	}

	// Transform single array values in scalars.
	params := make(map[string]any)
	for k, v := range values {
		switch len(v) {
		case 0:
			params[k] = nil
		case 1:
			var iface any
			if err = yaml.Unmarshal([]byte(v[0]), &iface); err != nil {
				return nil, err
			}
			params[k] = iface
		default:
			// It is a slice add element by element
			elemSlice := make([]any, 0, len(v))
			for _, elem := range v {
				var iface any
				if err = yaml.Unmarshal([]byte(elem), &iface); err != nil {
					return nil, err
				}
				elemSlice = append(elemSlice, iface)
			}
			params[k] = elemSlice
		}
	}
	return confmap.NewFromStringMap(params), err
}

// osExpandEnv replicate the internal behavior of os.ExpandEnv when handling env
// vars updating the buffer accordingly.
func osExpandEnv(buf []byte, name string, w int) []byte {
	switch {
	case name == "" && w > 0:
		// Encountered invalid syntax; eat the
		// characters.
	case name == "" || name == "$":
		// Valid syntax, but $ was not followed by a
		// name. Leave the dollar character untouched.
		buf = append(buf, expandPrefixChar)
	default:
		buf = append(buf, os.Getenv(name)...)
	}

	return buf
}

// scanToClosingBracket consumes everything until a closing bracket '}' following the
// same logic of function getShellName (os package, env.go) when handling environment
// variables with the "${<env_var>}" syntax. It returns the expression between brackets
// and the number of characters consumed from the original string.
func scanToClosingBracket(s string) (string, int) {
	for i := 1; i < len(s); i++ {
		if s[i] == '}' {
			if i == 1 {
				return "", 2 // Bad syntax; eat "${}"
			}
			return s[1:i], i + 1
		}
	}
	return "", 1 // Bad syntax; eat "${"
}

// getTokenName consumes characters until it has the name of either an environment
// variable or config source. It returns the name of the config source or environment
// variable and the number of characters consumed from the original string.
func getTokenName(s string) (string, int) {
	if len(s) > 0 && isShellSpecialVar(s[0]) {
		// Special shell character, treat it os.Expand function.
		return s[0:1], 1
	}

	var i int
	firstNameSepIdx := -1
	for i = 0; i < len(s); i++ {
		if isAlphaNum(s[i]) {
			// Continue while alphanumeric plus underscore.
			continue
		}

		if s[i] == typeAndNameSeparator && firstNameSepIdx == -1 {
			// If this is the first type name separator store the index and continue.
			firstNameSepIdx = i
			continue
		}

		// It is one of the following cases:
		// 1. End of string
		// 2. Reached a non-alphanumeric character, preceded by at most one
		//    typeAndNameSeparator character.
		break
	}

	if firstNameSepIdx != -1 && (i >= len(s) || s[i] != configSourceNameDelimChar) {
		// Found a second non alpha-numeric character before the end of the string
		// but it is not the config source delimiter. Use the name until the first
		// name delimiter.
		return s[:firstNameSepIdx], firstNameSepIdx
	}

	return s[:i], i
}

// Below are helper functions used by os.Expand, copied without changes from original sources (env.go).

// isShellSpecialVar reports whether the character identifies a special
// shell variable such as $*.
func isShellSpecialVar(c uint8) bool {
	switch c {
	case '*', '#', '$', '@', '!', '?', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

// isAlphaNum reports whether the byte is an ASCII letter, number, or underscore
func isAlphaNum(c uint8) bool {
	return c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func MergeCloseFuncs(closeFuncs []confmap.CloseFunc) confmap.CloseFunc {
	if len(closeFuncs) == 0 {
		return nil
	}
	if len(closeFuncs) == 1 {
		return closeFuncs[0]
	}
	return func(ctx context.Context) error {
		var errs error
		for _, closeFunc := range closeFuncs {
			if closeFunc != nil {
				errs = multierr.Append(errs, closeFunc(ctx))
			}
		}
		return errs
	}
}
