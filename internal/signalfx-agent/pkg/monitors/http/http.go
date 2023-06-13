package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/sirupsen/logrus"

	"github.com/signalfx/signalfx-agent/pkg/core/common/auth"
	"github.com/signalfx/signalfx-agent/pkg/core/common/httpclient"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

func init() {

	monitors.Register(&monitorMetadata, func() interface{} { return &Monitor{monitorName: monitorMetadata.MonitorType} }, &Config{})
}

// Config for this monitor
type Config struct {
	config.MonitorConfig `yaml:",inline" singleInstance:"false" acceptsEndpoints:"true"`
	// Host/IP to monitor
	Host string `yaml:"host"`
	// Port of the HTTP server to monitor
	Port uint16 `yaml:"port"`
	// HTTP path to use in the test request
	Path string `yaml:"path"`

	httpclient.HTTPConfig `yaml:",inline"`
	// Optional HTTP request body as string like '{"foo":"bar"}'
	RequestBody string `yaml:"requestBody"`
	// Do not follow redirect.
	NoRedirects bool `yaml:"noRedirects" default:"false"`
	// HTTP request method to use.
	Method string `yaml:"method" default:"GET"`
	// DEPRECATED: list of HTTP URLs to monitor. Use `host`/`port`/`useHTTPS`/`path` instead.
	URLs []string `yaml:"urls"`
	// Optional Regex to match on URL(s) response(s).
	Regex string `yaml:"regex"`
	// Desired code to match for URL(s) response(s).
	DesiredCode int `yaml:"desiredCode" default:"200"`
	// Add `redirect_url` dimension which could differ from `url` when redirection is followed.
	AddRedirectURL bool `yaml:"addRedirectURL" default:"false"`
}

// Monitor that collect metrics
type Monitor struct {
	Output types.FilteringOutput
	cancel context.CancelFunc
	//ctx         context.Context
	logger      logrus.FieldLogger
	conf        *Config
	monitorName string
	regex       *regexp.Regexp
	URLs        []*url.URL
}

// Configure and kick off internal metric collection
func (m *Monitor) Configure(conf *Config) (err error) {
	m.conf = conf
	m.logger = logrus.WithFields(logrus.Fields{"monitorType": m.monitorName, "monitorID": m.conf.MonitorID})
	// Ignore certificate error which will be checked after
	m.conf.SkipVerify = true

	if m.conf.Regex != "" {
		// Compile regex
		m.regex, err = regexp.Compile(m.conf.Regex)
		if err != nil {
			m.logger.WithError(err).Error("failed to compile regular expression")
		}
	}

	// Start the metric gathering process here
	var ctx context.Context
	ctx, m.cancel = context.WithCancel(context.Background())

	if m.conf.Host != "" {
		if m.conf.Port == 0 {
			m.conf.Port = m.conf.DefaultPort()
		}
		clientURL, err := m.normalizeURL(fmt.Sprintf("%s://%s:%d%s", m.conf.Scheme(), m.conf.Host, m.conf.Port, m.conf.Path))
		if err != nil {
			m.logger.WithError(err).Error("error configuring url from http client, ignore it")
		} else {
			m.URLs = append(m.URLs, clientURL)
		}
	} else {
		// always try https if available. This is for backwards compat with deprecated URLs.
		m.conf.UseHTTPS = true
	}
	for _, site := range m.conf.URLs {
		// add http scheme if not explicitly set
		if !strings.HasPrefix(site, "http") {
			site = fmt.Sprintf("http://%s", site)
		}
		stringURL, err := m.normalizeURL(site)
		if err != nil {
			m.logger.WithField("url", site).WithError(err).Error("error configuring url from list, ignore it")
			continue
		}
		m.URLs = append(m.URLs, stringURL)
	}

	utils.RunOnInterval(ctx, func() {
		// get stats for each website
		for _, site := range m.URLs {
			logger := m.logger.WithFields(logrus.Fields{"url": site.String()})

			dps, redirectURL, err := m.getHTTPStats(site, logger)
			if err == nil {
				if redirectURL.Scheme == "https" {
					tlsDps, err := m.getTLSStats(redirectURL, logger)
					if err == nil {
						dps = append(dps, tlsDps...)
					} else {
						logger.WithError(err).Error("Failed gathering TLS stats")
					}
				}
			} else {
				logger.WithError(err).Error("Failed gathering all HTTP stats, ignore TLS stats and push what we've successfully collected")
			}

			for i := range dps {
				dps[i].Dimensions["url"] = site.String()
			}

			if m.conf.AddRedirectURL && !m.conf.NoRedirects {
				query := redirectURL.RawQuery
				if query != "" {
					query = "?" + query
				}
				normalizedURL, _ := m.normalizeURL(fmt.Sprintf("%s://%s:%s%s%s", redirectURL.Scheme, redirectURL.Hostname(), redirectURL.Port(), redirectURL.Path, query))
				if site.String() != normalizedURL.String() {
					logger.WithField("redirect_url", normalizedURL.String()).Debug("URL redirected")
					for i := range dps {
						dps[i].Dimensions["redirect_url"] = normalizedURL.String()
					}
				}
			}

			m.Output.SendDatapoints(dps...)
		}
	}, time.Duration(conf.IntervalSeconds)*time.Second)

	return nil
}

// Shutdown the monitor
func (m *Monitor) Shutdown() {
	// Stop any long-running go routines here
	if m.cancel != nil {
		m.cancel()
	}
}

func (m *Monitor) normalizeURL(site string) (normalizedURL *url.URL, err error) {
	stringURL, err := url.Parse(site)
	if err != nil {
		return
	}
	host := stringURL.Hostname()
	port := stringURL.Port()
	// for deprecated URLs only, set default port if not explicitly set
	if host == stringURL.Host {
		port = "80"
	}
	// keep port only if custom, hide default
	if port != "80" && port != "443" {
		host = fmt.Sprintf("%s:%s", host, port)
	}
	path := stringURL.Path
	if path == "" {
		path = "/"
	}
	query := stringURL.RawQuery
	if query != "" {
		query = "?" + query
	}
	normalizedURL, err = url.Parse(fmt.Sprintf("%s://%s%s%s", stringURL.Scheme, host, path, query))
	if err != nil {
		return
	}
	return
}

func (m *Monitor) getTLSStats(site *url.URL, logger *logrus.Entry) (dps []*datapoint.Datapoint, err error) {
	// use as an fmt.Stringer
	host := site.Hostname()
	port := site.Port()
	serverName := m.conf.SNIServerName

	var valid int64 = 1
	var secondsLeft float64

	if port == "" {
		port = "443"
	}

	if serverName == "" {
		serverName = host
	}

	dimensions := map[string]string{
		"server_name":     host,
		"sni_server_name": serverName,
	}

	ipConn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		logger.WithError(err).Error("connection failed to host during TLS stat collection")
		return
	}
	defer ipConn.Close()

	tlsCfg := &tls.Config{
		ServerName: serverName,
	}

	if _, err := auth.TLSConfig(tlsCfg, m.conf.CACertPath, m.conf.ClientCertPath, m.conf.ClientKeyPath); err != nil {
		return nil, err
	}

	conn := tls.Client(ipConn, tlsCfg)
	if err != nil {
		return
	}
	defer conn.Close()

	err = conn.Handshake()
	if err != nil {
		logger.WithError(err).Debug("cert verification failed during handshake")
		valid = 0
	} else {
		cert := conn.ConnectionState().PeerCertificates[0]
		secondsLeft = time.Until(cert.NotAfter).Seconds()
	}

	dps = append(dps,
		datapoint.New(httpCertExpiry, dimensions, datapoint.NewFloatValue(secondsLeft), datapoint.Gauge, time.Time{}),
		datapoint.New(httpCertValid, dimensions, datapoint.NewIntValue(valid), datapoint.Gauge, time.Time{}))

	return dps, nil
}

func (m *Monitor) getHTTPStats(site fmt.Stringer, logger *logrus.Entry) (dps []*datapoint.Datapoint, redirectURL *url.URL, err error) {
	// do not suggest fmt.Stringer
	// Init http client
	client, err := m.conf.HTTPConfig.Build()
	if err != nil {
		return
	}

	// Init body if applicable
	var body io.Reader
	if m.conf.RequestBody != "" {
		body = strings.NewReader(m.conf.RequestBody)
	}

	if m.conf.NoRedirects {
		logger.Debug("Do not follow redirects")
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	req, err := http.NewRequest(m.conf.Method, site.String(), body)
	if err != nil {
		return
	}

	// override excluded headers, see:
	// https://github.com/golang/go/blob/cad6d1fef5147d31e94ee83934c8609d3ad150b7/src/net/http/request.go#L92
	if len(m.conf.HTTPHeaders) > 0 {
		for key, val := range m.conf.HTTPHeaders {
			if strings.EqualFold(key, "host") {
				req.Host = val
				continue
			}
			if strings.EqualFold(key, "content-length") {
				if contentLenght, err := strconv.Atoi(val); err == nil {
					req.ContentLength = int64(contentLenght)
				}
				continue
			}
		}
	}

	// starts timer
	now := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	responseTime := time.Since(now).Seconds()

	redirectURL = resp.Request.URL

	dimensions := map[string]string{
		"method":   m.conf.Method,
		"req_host": req.Host,
	}

	statusCode := int64(resp.StatusCode)

	var matchCode int64 = 0
	if statusCode == int64(m.conf.DesiredCode) {
		matchCode = 1
	}

	dps = append(dps,
		datapoint.New(httpResponseTime, dimensions, datapoint.NewFloatValue(responseTime), datapoint.Gauge, time.Time{}),
		datapoint.New(httpStatusCode, dimensions, datapoint.NewIntValue(statusCode), datapoint.Gauge, time.Time{}),
		datapoint.New(httpCodeMatched, dimensions, datapoint.NewIntValue(matchCode), datapoint.Gauge, time.Time{}),
	)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.WithError(err).Error("could not parse body response")
	} else {
		dps = append(dps, datapoint.New(httpContentLength, dimensions, datapoint.NewIntValue(int64(len(bodyBytes))), datapoint.Gauge, time.Time{}))

		if m.conf.Regex != "" {
			var matchRegex int64 = 0
			if m.regex.Match(bodyBytes) {
				matchRegex = 1
			}
			dps = append(dps, datapoint.New(httpRegexMatched, dimensions, datapoint.NewIntValue(matchRegex), datapoint.Gauge, time.Time{}))
		}
	}
	return dps, redirectURL, err
}
