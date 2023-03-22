package haproxy

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	log "github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

func init() {
	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{} }, &Config{})
}

// Monitor for reporting HAProxy stats.
type Monitor struct {
	Output types.Output
	cancel context.CancelFunc
	ctx    context.Context
	logger log.FieldLogger
}

// Map of proxies to monitor
type proxies map[string]bool

// Config for this monitor
func (m *Monitor) Configure(conf *Config) (err error) {
	m.logger = log.WithFields(log.Fields{"monitorType": monitorType, "monitorID": conf.MonitorID})
	m.ctx, m.cancel = context.WithCancel(context.Background())
	url, err := url.Parse(conf.ScrapeURL())
	if err != nil {
		return fmt.Errorf("cannot parse url %s status. %v", conf.ScrapeURL(), err)
	}
	pxs := proxies{}
	for _, p := range conf.Proxies {
		switch strings.ToLower(strings.TrimSpace(p)) {
		case "frontend":
			pxs["FRONTEND"] = true
		case "backend":
			pxs["BACKEND"] = true
		default:
			pxs[p] = true
		}
	}
	type funcs []func(*Config, proxies) []*datapoint.Datapoint
	var fetchFuncs funcs
	switch url.Scheme {
	case "http", "https", "file":
		fetchFuncs = funcs{m.statsHTTP}
	case "unix":
		fetchFuncs = funcs{m.statsSocket, m.infoSocket}
	default:
		return fmt.Errorf("unsupported url scheme:%q", url.Scheme)
	}
	interval := time.Duration(conf.IntervalSeconds) * time.Second
	utils.RunOnInterval(m.ctx, func() {
		ctx, cancel := context.WithTimeout(m.ctx, interval)
		defer cancel()
		var wg sync.WaitGroup
		chs := make([]chan []*datapoint.Datapoint, 0)
		for _, fn := range fetchFuncs {
			fn := fn
			ch := make(chan []*datapoint.Datapoint, 1)
			wg.Add(1)
			go func() {
				defer close(ch)
				defer wg.Done()
				select {
				case <-ctx.Done():
					m.logger.Error(ctx.Err())
					return
				case ch <- fn(conf, pxs):
					return
				}
			}()
			chs = append(chs, ch)
		}
		wg.Wait()
		for _, ch := range chs {
			dps := <-ch
			for i := range dps {
				dps[i].Dimensions["plugin"] = "haproxy"
			}
			m.Output.SendDatapoints(dps...)
		}
	}, interval)
	return nil
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}
