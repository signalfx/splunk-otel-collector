package prometheusexporter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"

	"github.com/signalfx/signalfx-agent/pkg/core/common/auth"
	"github.com/signalfx/signalfx-agent/pkg/core/common/httpclient"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

func init() {
	RegisterMonitor(monitorMetadata)
}

// Config for this monitor
type Config struct {
	config.MonitorConfig  `yaml:",inline" acceptsEndpoints:"true"`
	httpclient.HTTPConfig `yaml:",inline"`

	// Host of the exporter
	Host string `yaml:"host" validate:"required"`
	// Port of the exporter
	Port uint16 `yaml:"port" validate:"required"`

	// Use pod service account to authenticate.
	UseServiceAccount bool `yaml:"useServiceAccount"`

	// Path to the metrics endpoint on the exporter server, usually `/metrics`
	// (the default).
	MetricPath string `yaml:"metricPath" default:"/metrics"`

	// Control the log level to use if a scrape failure occurs when scraping
	// a target. Modifying this configuration is useful for less stable
	// targets. All logrus log levels are supported.
	ScrapeFailureLogLevel    string `yaml:"scrapeFailureLogLevel" default:"error"`
	scrapeFailureLogrusLevel logrus.Level

	// Send all the metrics that come out of the Prometheus exporter without
	// any filtering.  This option has no effect when using the prometheus
	// exporter monitor directly since there is no built-in filtering, only
	// when embedding it in other monitors.
	SendAllMetrics bool `yaml:"sendAllMetrics"`
}

func (c *Config) Validate() error {
	l, err := logrus.ParseLevel(c.ScrapeFailureLogLevel)
	if err != nil {
		return err
	}
	c.scrapeFailureLogrusLevel = l
	return nil
}

func (c *Config) GetExtraMetrics() []string {
	// Maintain backwards compatibility with the config flag that existing
	// prior to the new filtering mechanism.
	if c.SendAllMetrics {
		return []string{"*"}
	}
	return nil
}

var _ config.ExtraMetrics = &Config{}

// Monitor for prometheus exporter metrics
type Monitor struct {
	Output types.Output
	// Optional set of metric names that will be sent by default, all other
	// metrics derived from the exporter being dropped.
	IncludedMetrics map[string]bool
	// Extra dimensions to add in addition to those specified in the config.
	ExtraDimensions map[string]string
	// If true, IncludedMetrics is ignored and everything is sent.
	SendAll bool

	monitorName string
	logger      logrus.FieldLogger
	cancel      func()
}

type fetcher func() (io.ReadCloser, expfmt.Format, error)

// Configure the monitor and kick off volume metric syncing
func (m *Monitor) Configure(conf *Config) error {
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": m.monitorName, "monitorID": conf.MonitorID})

	var bearerToken string

	if conf.UseServiceAccount {
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return err
		}
		bearerToken = restConfig.BearerToken
		if bearerToken == "" {
			return errors.New("bearer token was empty")
		}
	}

	client, err := conf.HTTPConfig.Build()
	if err != nil {
		return err
	}

	if bearerToken != "" {
		client.Transport = &auth.TransportWithToken{
			RoundTripper: client.Transport,
			Token:        bearerToken,
		}
	}

	url := fmt.Sprintf("%s://%s:%d%s", conf.Scheme(), conf.Host, conf.Port, conf.MetricPath)

	fetch := func() (io.ReadCloser, expfmt.Format, error) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, expfmt.FmtUnknown, err
		}

		resp, err := client.Do(req) // nolint:bodyclose  // We do actually close it after it is returned
		if err != nil {
			return nil, expfmt.FmtUnknown, err
		}

		if resp.StatusCode != 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, expfmt.FmtUnknown, fmt.Errorf("prometheus exporter at %s returned status %d: %s", url, resp.StatusCode, string(body))
		}

		return resp.Body, expfmt.ResponseFormat(resp.Header), nil
	}

	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())
	utils.RunOnInterval(ctx, func() {
		dps, err := fetchPrometheusMetrics(fetch)
		if err != nil {
			// The default log level is error, users can configure which level to use
			m.logger.WithError(err).Log(conf.scrapeFailureLogrusLevel, "Could not get prometheus metrics")
			return
		}

		m.Output.SendDatapoints(dps...)
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

func fetchPrometheusMetrics(fetch fetcher) ([]*datapoint.Datapoint, error) {
	metricFamilies, err := doFetch(fetch)
	if err != nil {
		return nil, err
	}

	var dps []*datapoint.Datapoint
	for i := range metricFamilies {
		dps = append(dps, convertMetricFamily(metricFamilies[i])...)
	}
	return dps, nil
}

func doFetch(fetch fetcher) ([]*dto.MetricFamily, error) {
	body, expformat, err := fetch()
	if err != nil {
		return nil, err
	}
	defer body.Close()
	var decoder expfmt.Decoder
	// some "text" responses are missing \n from the last line
	if expformat != expfmt.FmtProtoDelim {
		decoder = expfmt.NewDecoder(io.MultiReader(body, strings.NewReader("\n")), expformat)
	} else {
		decoder = expfmt.NewDecoder(body, expformat)
	}

	var mfs []*dto.MetricFamily

	for {
		var mf dto.MetricFamily
		err := decoder.Decode(&mf)

		if err == io.EOF {
			return mfs, nil
		} else if err != nil {
			return nil, err
		}

		mfs = append(mfs, &mf)
	}
}

// Shutdown stops the metric sync
func (m *Monitor) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}
}

// RegisterMonitor is a helper for other monitors that simply wrap prometheusexporter.
func RegisterMonitor(meta monitors.Metadata) {
	monitors.Register(&meta, func() interface{} {
		return &Monitor{monitorName: meta.MonitorType}
	},
		&Config{})
}
