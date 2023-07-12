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
