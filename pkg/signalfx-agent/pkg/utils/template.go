package utils

import (
	"bytes"
	"text/template"
)

// RenderSimpleTemplate processes a simple, self-contained template with a
// given context and returns the final result.
func RenderSimpleTemplate(tmpl string, context interface{}) (string, error) {
	template, err := template.New("nested").Parse(tmpl)
	if err != nil {
		return "", err
	}

	out := bytes.Buffer{}
	// fill in any templates with the whole config struct passed into this method
	err = template.Option("missingkey=error").Execute(&out, context)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}
