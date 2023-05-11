// Package zookeeper contains the logic for using Zookeeper as a config source.
// Currently globbing only works if it is a suffix to the path.
package zookeeper

import (
	"errors"
	"fmt"
	"hash/crc64"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"github.com/signalfx/defaults"
	"github.com/signalfx/signalfx-agent/pkg/core/config/types"
	log "github.com/sirupsen/logrus"
)

type zkConfigSource struct {
	conn      *zk.Conn
	endpoints []string
	timeout   time.Duration
	table     *crc64.Table
}

// Config is used to configure the Zookeeper client
type Config struct {
	// A list of Zookeeper servers to use for the client
	Endpoints []string `yaml:"endpoints"`
	// Client timeout
	TimeoutSeconds uint `yaml:"timeoutSeconds" default:"10"`
}

// New creates a new Zookeeper remote config source from the target config
func (c *Config) New() (types.ConfigSource, error) {
	return New(c), nil
}

// Validate the config
func (c *Config) Validate() error {
	return nil
}

var _ types.ConfigSourceConfig = &Config{}

// New creates a new Zookeeper config source with the given config.  All gets
// and watches will use the same client.
func New(conf *Config) types.ConfigSource {
	_ = defaults.Set(conf)
	return &zkConfigSource{
		endpoints: conf.Endpoints,
		table:     crc64.MakeTable(crc64.ECMA),
		timeout:   time.Duration(conf.TimeoutSeconds) * time.Second,
	}
}

func (z *zkConfigSource) ensureConnection() error {
	if z.conn != nil && z.conn.State() != zk.StateDisconnected {
		return nil
	}

	conn, _, err := zk.Connect(z.endpoints, z.timeout, zk.WithLogInfo(false))
	if err != nil {
		return err
	}
	z.conn = conn
	return nil
}

// ErrBadGlob gets returned when the globbing in a path is invalid
var ErrBadGlob = errors.New("zookeeper only supports globs in the last path segment")

func isGlob(path string) (string, bool, error) {
	firstGlobIndex := strings.IndexAny(path, "*?")
	if strings.HasSuffix(path, "/*") {
		if firstGlobIndex != len(path)-1 {
			return "", false, ErrBadGlob
		}
		return strings.TrimSuffix(path, "/*"), true, nil
	}
	if firstGlobIndex != -1 {
		return "", false, ErrBadGlob
	}
	return "", false, nil
}

func (z *zkConfigSource) Name() string {
	return "zookeeper"
}

func (z *zkConfigSource) Get(path string) (map[string][]byte, uint64, error) {
	content, version, _, err := z.getNodes(path, false)
	return content, version, err
}

// The Zookeeper go lib is really not amenable to the pattern we use for
// ConfigSource since it doesn't have any concept of a global index counter
// that can be used to ensure updates aren't missed, and it also doesn't have
// the ability to watch child nodes recursively.  Therefore, for now we just
// limit globbing to asterisks at the very end of a node.  If we need more
// complex globbing, consider something like
// https://github.com/kelseyhightower/confd/blob/master/backends/zookeeper/client.go
func (z *zkConfigSource) getNodes(path string, watch bool) (map[string][]byte, uint64, []<-chan zk.Event, error) {
	if err := z.ensureConnection(); err != nil {
		return nil, 0, nil, err
	}

	contentMap := make(map[string][]byte)
	var sums string
	var nodes []string
	var events []<-chan zk.Event

	prefix, globbed, err := isGlob(path)
	if err != nil {
		return nil, 0, nil, err
	}
	if globbed {
		var err error
		var parentEvents <-chan zk.Event
		if watch {
			nodes, _, parentEvents, err = z.conn.ChildrenW(prefix)
		} else {
			nodes, _, err = z.conn.Children(prefix)
		}

		if err != nil {
			return nil, 0, nil, err
		}
		if parentEvents != nil {
			events = append(events, parentEvents)
		}
	} else {
		nodes = []string{path}
	}

	sort.Strings(nodes)

	for _, n := range nodes {
		var content []byte
		var stat *zk.Stat
		var err error
		var nodeEvents <-chan zk.Event

		fullPath := filepath.Join(prefix, n)

		if watch {
			content, stat, nodeEvents, err = z.conn.GetW(fullPath)
		} else {
			content, stat, err = z.conn.Get(fullPath)
		}
		if err != nil {
			return nil, 0, nil, err
		}

		contentMap[n] = content
		sums = fmt.Sprintf("%s:%s:%d", sums, fullPath, stat.Version)

		if nodeEvents != nil {
			events = append(events, nodeEvents)
		}
	}

	return contentMap, crc64.Checksum([]byte(sums), z.table), events, nil
}

func (z *zkConfigSource) WaitForChange(path string, version uint64, stop <-chan struct{}) error {
	_, newVersion, events, err := z.getNodes(path, true)
	if err != nil {
		return err
	}

	if version != newVersion {
		return nil
	}

	cases := []reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(stop),
		},
	}
	for _, ch := range events {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		})
	}

	for {
		log.Debugf("Waiting for ZK change at %s with %d nodes", path, len(cases)-1)
		chosen, _, _ := reflect.Select(cases)
		// Stop channel is the first
		if chosen == 0 {
			return nil
		}
		log.Debugf("ZK path %s changed", path)

		// Get the data again and compare versions so that we don't have false
		// positives.
		_, newVersion, err := z.Get(path)
		if err != nil {
			return err
		}
		if newVersion != version {
			return nil
		}
	}
}
