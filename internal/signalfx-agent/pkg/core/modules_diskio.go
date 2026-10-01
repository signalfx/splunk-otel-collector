//go:build !windows && !aix && !freebsd && !solaris

package core

import _ "github.com/signalfx/signalfx-agent/pkg/monitors/diskio"
