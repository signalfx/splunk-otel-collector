package config

import (
	"fmt"
	"reflect"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/davecgh/go-spew/spew"
	"github.com/signalfx/defaults"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	log "github.com/sirupsen/logrus"
)

// DecodeExtraConfigStrict will pull out any config values from 'in' and put them
// on the 'out' struct, returning an error if anything in 'in' isn't in 'out'.
func DecodeExtraConfigStrict(in CustomConfigurable, out interface{}) error {
	return DecodeExtraConfig(in, out, true)
}

// DecodeExtraConfig will pull out the OtherConfig values from both
// ObserverConfig and MonitorConfig and decode them to a struct that is
// provided in the `out` arg.  Whether all fields have to be in 'out' is
// determined by the 'strict' flag.  Any errors decoding will cause `out` to be nil.
func DecodeExtraConfig(in CustomConfigurable, out interface{}, strict bool) error {
	pkgPaths := strings.Split(reflect.Indirect(reflect.ValueOf(out)).Type().PkgPath(), "/")

	extra, err := in.ExtraConfig()
	if err != nil {
		return err
	}

	otherYaml, err := yaml.Marshal(extra)
	if err != nil {
		return err
	}

	if strict {
		err = yaml.UnmarshalStrict(otherYaml, out)
	} else {
		// Preserve any special "catch-all" config that is typed with the
		// special AdditionalConfig type.  This is needed because many config
		// structs will be Unmarshaled to multiple times from various sources,
		// and any inline struct fields will be overwritten on subsequent
		// Unmarshals.
		var additionalConf AdditionalConfig
		additionalConfigField := utils.FindFirstFieldOfType(out, reflect.TypeOf((AdditionalConfig)(nil)))
		if additionalConfigField.IsValid() {
			additionalConf = additionalConfigField.Interface().(AdditionalConfig)
		}

		err = yaml.Unmarshal(otherYaml, out)

		for k, v := range additionalConf {
			additionalConfigField.Interface().(AdditionalConfig)[k] = v
		}
	}
	if err != nil {
		return err
	}

	if err := defaults.Set(out); err != nil {
		log.WithFields(log.Fields{
			"package": pkgPaths[len(pkgPaths)-1],
			"error":   err,
			"out":     spew.Sdump(out),
		}).Error("Could not set defaults on module-specific config")
		return err
	}
	return nil
}

// FillInConfigTemplate takes a config template value that a monitor/observer
// provided and fills it in dynamically from the provided conf
func FillInConfigTemplate(embeddedFieldName string, configTemplate interface{}, conf CustomConfigurable) error {
	templateValue := reflect.ValueOf(configTemplate)
	pkg := templateValue.Type().PkgPath()

	if templateValue.Kind() != reflect.Ptr || templateValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config template must be a pointer to a struct, got %s of kind/type %s/%s",
			pkg, templateValue.Kind(), templateValue.Type())
	}

	embeddedField := templateValue.Elem().FieldByName(embeddedFieldName)
	if !embeddedField.IsValid() {
		return fmt.Errorf("could not find field %s in config.  Available fields: %v",
			embeddedFieldName, utils.GetStructFieldNames(templateValue))
	}
	embeddedField.Set(reflect.Indirect(reflect.ValueOf(conf)))

	return DecodeExtraConfigStrict(conf, configTemplate)
}

// CallConfigure will call the Configure method on an observer or monitor with
// a `conf` object, typed to the correct type.  This allows monitors/observers
// to set the type of the config object to their own config and not have to
// worry about casting or converting.
func CallConfigure(instance, conf interface{}) error {
	instanceVal := reflect.ValueOf(instance)
	_type := instanceVal.Type().PkgPath()

	confVal := reflect.ValueOf(conf)

	method := instanceVal.MethodByName("Configure")
	if !method.IsValid() {
		return fmt.Errorf("no Configure method found for type %s", _type)
	}

	if method.Type().NumIn() != 1 {
		return fmt.Errorf("configure method of %s should take exactly one argument that matches "+
			"the type of the config template provided in the Register function! It has %d arguments.",
			_type, method.Type().NumIn())
	}

	errorIntf := reflect.TypeOf((*error)(nil)).Elem()
	if method.Type().NumOut() != 1 || !method.Type().Out(0).Implements(errorIntf) {
		return fmt.Errorf("configure method for type %s should return an error", _type)
	}

	ret := method.Call([]reflect.Value{confVal})[0]
	if ret.IsNil() {
		return nil
	}
	return ret.Interface().(error)
}
