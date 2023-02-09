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

package collectd

import (
	"bytes"
	"text/template"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

// RenderValue renders a template value
func RenderValue(templateText string, context interface{}) (string, error) {
	if templateText == "" {
		return "", nil
	}

	template, err := template.New("nested").Parse(templateText)
	if err != nil {
		return "", err
	}

	out := bytes.Buffer{}
	err = template.Option("missingkey=error").Execute(&out, context)
	if err != nil {
		log.WithFields(log.Fields{
			"template": templateText,
			"error":    err,
			"context":  spew.Sdump(context),
		}).Error("Could not render nested config template")
		return "", err
	}

	return out.String(), nil
}
