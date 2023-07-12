package file

import (
	"fmt"
	"hash/crc64"
	"io/ioutil"
	"path/filepath"
	"sort"
	"time"

	"github.com/signalfx/signalfx-agent/pkg/core/config/types"
)

// Config for the file-based config source
type Config struct {
	// How often to poll files (in seconds) to test for changes.  There are so
	// many edge cases that break inotify that it is more robust to simply poll
	// files than rely on that.
	// This option is not subject to watching and changes to it will require an
	// agent restart.
	PollRateSeconds int `yaml:"pollRateSeconds" default:"5"`
}

// New creates a new file remote config source from the target config
func (c *Config) New() (types.ConfigSource, error) {
	return New(time.Duration(c.PollRateSeconds) * time.Second), nil
}

// Validate the config
func (c *Config) Validate() error {
	return nil
}

var _ types.ConfigSourceConfig = &Config{}

type fileConfigSource struct {
	table        *crc64.Table
	pollInterval time.Duration
}

// New makes a new fileConfigSource with the given config
func New(pollInterval time.Duration) types.ConfigSource {
	return &fileConfigSource{
		table:        crc64.MakeTable(crc64.ECMA),
		pollInterval: pollInterval,
	}
}

func (fcs *fileConfigSource) Name() string {
	return "file"
}

func (fcs *fileConfigSource) Get(path string) (map[string][]byte, uint64, error) {
	matches, err := filepath.Glob(path)
	if err != nil {
		return nil, 0, err
	}
	if len(matches) == 0 {
		return nil, 0, types.NewNotFoundError("file(s) not found")
	}

	// sort so the checksum is consistent
	sort.Strings(matches)

	contentMap := make(map[string][]byte)
	var sums string

	for _, file := range matches {
		content, checksum, err := fcs.getFile(file)
		if err != nil {
			return nil, 0, err
		}
		sums = fmt.Sprintf("%s:%d", sums, checksum)
		contentMap[file] = content
	}

	return contentMap, crc64.Checksum([]byte(sums), fcs.table), nil
}

func (fcs *fileConfigSource) getFile(path string) ([]byte, uint64, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}

	checksum := crc64.Checksum(content, fcs.table)
	return content, checksum, nil
}

// WaitForChange polls the file for changes.  There are too many edge cases and
// subtleties (e.g. symlinks, remote mounts, etc.) when using inotify (or the
// wrapping golang library, fsnotify), so the simplest and most robust thing is
// to just poll the file and see if it has changed.
func (fcs *fileConfigSource) WaitForChange(path string, version uint64, stop <-chan struct{}) error {
	ticker := time.NewTicker(fcs.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			_, newVersion, err := fcs.Get(path)

			if err != nil {
				return err
			}
			if newVersion != version {
				return nil
			}
		}
	}
}
