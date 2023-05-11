package consul

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/consul/api"
	"github.com/signalfx/signalfx-agent/pkg/core/config/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

type consulConfigSource struct {
	client *api.Client
	kv     *api.KV
}

// Config for the consul client
type Config struct {
	// A Consul server URL
	Endpoint string `yaml:"endpoint"`
	// An optional username to use when connecting
	Username string `yaml:"username"`
	// An optional password to use when connecting
	Password string `yaml:"password" neverLog:"true"`
	// An authentication token, if needed
	Token string `yaml:"token" neverLog:"true"`
	// The Consul datacenter to use
	Datacenter string `yaml:"datacenter"`
}

// New creates a new Consul remote config source from the target config
func (c *Config) New() (types.ConfigSource, error) {
	return New(c)
}

// Validate the config
func (c *Config) Validate() error {
	return nil
}

var _ types.ConfigSourceConfig = &Config{}

// New creates a new consul ConfigSource
func New(conf *Config) (types.ConfigSource, error) {
	var httpAuth *api.HttpBasicAuth
	if conf.Username != "" || conf.Password != "" {
		httpAuth = &api.HttpBasicAuth{
			Username: conf.Username,
			Password: conf.Password,
		}
	}

	c, err := api.NewClient(&api.Config{
		Address:    conf.Endpoint,
		HttpAuth:   httpAuth,
		Token:      conf.Token,
		Datacenter: conf.Datacenter,
	})
	if err != nil {
		return nil, err
	}

	kv := c.KV()

	return &consulConfigSource{
		client: c,
		kv:     kv,
	}, nil
}

func (c *consulConfigSource) Name() string {
	return "consul"
}

func (c *consulConfigSource) Get(path string) (map[string][]byte, uint64, error) {
	// Take off leading / from consul paths since they aren't official.
	prefix, g, _, err := types.PrefixAndGlob(strings.TrimPrefix(path, "/"))
	if err != nil {
		return nil, 0, err
	}

	pairs, meta, err := c.kv.List(prefix, nil)

	if err != nil {
		return nil, 0, err
	}

	contentMap := make(map[string][]byte)
	for _, pair := range pairs {
		if !g.Match(pair.Key) {
			continue
		}
		contentMap[pair.Key] = pair.Value
	}

	return contentMap, meta.LastIndex, nil
}

func (c *consulConfigSource) WaitForChange(path string, version uint64, stop <-chan struct{}) error {
	prefix, g, _, err := types.PrefixAndGlob(strings.TrimPrefix(path, "/"))
	if err != nil {
		return err
	}

	event := make(chan error)
	// The consul client doesn't provide any way of cancelling wait
	// requests, so do the List call in a separate goroutine and have it
	// send events through the event channel.  Make the watch finish after
	// a certain amount of time so if the path is no longer watched in the
	// agent, it won't keep a connection open indefinitely.
	// This technique is inspired by the Consul backend of
	// https://github.com/kelseyhightower/confd
	go func() {
		for {
			pairs, meta, err := c.kv.List(prefix, &api.QueryOptions{
				WaitIndex: version,
				WaitTime:  10 * time.Minute,
			})
			if utils.IsSignalChanClosed(stop) {
				return
			}
			if err != nil {
				event <- err
			}

			var anyMatch bool
			for _, p := range pairs {
				if g.Match(p.Key) {
					anyMatch = true
					break
				}
			}

			if anyMatch && meta.LastIndex > version {
				event <- nil
			}
		}
	}()

	select {
	case <-stop:
		return nil
	case err := <-event:
		log.Infof("Consul returned event %s for path %s", err, prefix)
		return err
	}
}
