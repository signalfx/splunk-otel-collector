package etcd2

import (
	"context"

	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/client/v2"

	"github.com/signalfx/signalfx-agent/pkg/core/config/types"
)

type etcd2ConfigSource struct {
	client *client.Client
	kapi   client.KeysAPI
}

// Config for an Etcd2 source
type Config struct {
	// A list of Etcd2 servers to use
	Endpoints []string `yaml:"endpoints"`
	// An optional username to use when connecting
	Username string `yaml:"username"`
	// An optional password to use when connecting
	Password string `yaml:"password" neverLog:"true"`
}

// New creates a new Etcd2 remote config source from the target config
func (c *Config) New() (types.ConfigSource, error) {
	return New(c)
}

// Validate the config
func (c *Config) Validate() error {
	return nil
}

var _ types.ConfigSourceConfig = &Config{}

// New creates a new etcd2 config source
func New(conf *Config) (types.ConfigSource, error) {
	c, err := client.New(client.Config{
		Endpoints: conf.Endpoints,
		Username:  conf.Username,
		Password:  conf.Password,
	})
	if err != nil {
		return nil, err
	}

	kapi := client.NewKeysAPI(c)

	return &etcd2ConfigSource{
		client: &c,
		kapi:   kapi,
	}, nil
}

func (e *etcd2ConfigSource) Name() string {
	return "etcd2"
}

func matchNodeKeys(node *client.Node, g glob.Glob, contentMap map[string][]byte) {
	if g.Match(node.Key) {
		contentMap[node.Key] = []byte(node.Value)
	}

	for _, n := range node.Nodes {
		log.Infof("Testing key %s", n.Key)
		if g.Match(n.Key) {
			contentMap[n.Key] = []byte(n.Value)
		}
		matchNodeKeys(n, g, contentMap)
	}
}

func (e *etcd2ConfigSource) Get(path string) (map[string][]byte, uint64, error) {
	prefix, g, isGlob, err := types.PrefixAndGlob(path)
	if err != nil {
		return nil, 0, err
	}

	resp, err := e.kapi.Get(context.Background(), prefix, &client.GetOptions{
		Recursive: isGlob,
	})

	if err != nil {
		if client.IsKeyNotFound(err) {
			return nil, 0, types.NewNotFoundError("etcd2 key not found")
		}
		return nil, 0, err
	}

	contentMap := make(map[string][]byte)
	matchNodeKeys(resp.Node, g, contentMap)

	return contentMap, resp.Index, nil
}

func (e *etcd2ConfigSource) WaitForChange(path string, version uint64, stop <-chan struct{}) error {
	prefix, g, isGlob, err := types.PrefixAndGlob(path)
	if err != nil {
		return err
	}

	watcher := e.kapi.Watcher(prefix, &client.WatcherOptions{
		AfterIndex: version,
		Recursive:  isGlob,
	})

	for {
		ctx, cancel := context.WithCancel(context.Background())
		watchDone := make(chan struct{})

		go func() {
			select {
			case <-watchDone:
				return
			case <-stop:
				cancel()
			}
		}()

		resp, err := watcher.Next(ctx)
		close(watchDone)
		cancel()

		if err != nil {
			return err
		}

		if g.Match(resp.Node.Key) && resp.Node.ModifiedIndex > version {
			return nil
		}
	}
}
