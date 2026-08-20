package diskqueuestorageextension

import (
	"errors"
	"time"
)

type Config struct {
	Path            string        `mapstructure:"path"`
	MaxBytesPerFile int64         `mapstructure:"max_bytes_per_file"`
	SyncEvery       int64         `mapstructure:"sync_every"`
	SyncTimeout     time.Duration `mapstructure:"sync_timeout"`
	CompactInterval time.Duration `mapstructure:"compact_interval"`
}

func (c *Config) Validate() error {
	if c.Path == "" {
		return errors.New("path must be a valid folder")
	}
	return nil
}
