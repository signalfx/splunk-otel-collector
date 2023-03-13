package structtags

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

const (
	copyToTag = "copyTo"
)

// CopyTo -
func CopyTo(ptr interface{}) error {

	v := reflect.ValueOf(ptr).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		if val := t.Field(i).Tag.Get(copyToTag); val != "-" && val != "" {
			var targets []string

			// initialize with tag value
			var groups = []string{val}

			// break apart targets and os commands
			if strings.Contains(val, ",GOOS=") {
				groups = strings.Split(val, ",GOOS=")
			}

			// break apart the targets
			targets = strings.Split(groups[0], ",")

			// check os eligibility
			eligibleOS := true
			if len(groups) == 2 {
				eligibleOS = isOSEligible(groups[1])
			}

			// if eligible copy the value to the targets
			if eligibleOS {
				for _, target := range targets {
					sourceField := v.Field(i)
					targetField := v.FieldByName(target)
					if targetField.CanSet() && sourceField.Kind() == targetField.Kind() {
						targetField.Set(v.Field(i))
					} else {
						return fmt.Errorf("unable to copy struct %v to target %s", sourceField, target)
					}
				}
			}
		}
	}
	return nil
}

// isOSEligible - determines if the os is eligible from the array of strings
func isOSEligible(osString string) bool {
	// if the os string is empty
	if osString == "" {
		return true
	}
	// check if the current os is explicitly excluded Ex. "!windows"
	if strings.Contains(osString, fmt.Sprintf("!%s", runtime.GOOS)) {
		return false
	}
	// check if the os is explicitly included Ex. "windows"
	if strings.Contains(osString, runtime.GOOS) {
		return true
	}
	// check for explicitly defined operating systems Ex. windows != "linux"
	operatingSystems := strings.Split(osString, ",")
	for _, f := range operatingSystems {
		if !strings.Contains(f, "!") {
			return false
		}
	}
	// any explicitly listed oses were exclusionary
	// and the runtime operating system doesn't match any of them
	return true
}
