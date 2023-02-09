// Copyright  Splunk, Inc.
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

package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	validator "gopkg.in/go-playground/validator.v9"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"
)

// Validatable should be implemented by config structs that want to provide
// validation when the config is loaded.
type Validatable interface {
	Validate() error
}

// ValidateCustomConfig for module-specific config ahead of time for a specific
// module configuration.  This way, the Configure method of modules will be
// guaranteed to receive valid configuration.  The module-specific
// configuration struct must implement the Validate method that returns a bool.
func ValidateCustomConfig(conf interface{}) error {
	if v, ok := conf.(Validatable); ok {
		return v.Validate()
	}
	return nil
}

// ValidateStruct uses the `validate` struct tags to do standard validation
func ValidateStruct(confStruct interface{}) error {
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")

	validate := validator.New()
	validate.RegisterTagNameFunc(utils.YAMLNameOfField)
	_ = en_translations.RegisterDefaultTranslations(validate, trans)
	err := validate.Struct(confStruct)
	if err != nil {
		if ves, ok := err.(validator.ValidationErrors); ok {
			var msgs []string
			for _, e := range ves {
				msgs = append(msgs, fmt.Sprintf("Validation error in field '%s': %s (got '%v')", e.Namespace(), e.Translate(trans), e.Value()))
			}
			return errors.New(strings.Join(msgs, "; "))
		}
		return err
	}
	return nil
}
