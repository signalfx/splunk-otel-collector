//go:build darwin
// +build darwin

package processlist

import "github.com/sirupsen/logrus"

type osCache struct {
}

func initOSCache() *osCache {
	return &osCache{}
}

func ProcessList(conf *Config, cache *osCache, logger logrus.FieldLogger) ([]*TopProcess, error) {
	return nil, nil
}
