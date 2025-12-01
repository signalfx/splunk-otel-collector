//go:build darwin

package processlist

import "github.com/sirupsen/logrus"

type osCache struct{}

func initOSCache() *osCache {
	return &osCache{}
}

func ProcessList(_ *Config, _ *osCache, _ logrus.FieldLogger) ([]*TopProcess, error) {
	return nil, nil
}
