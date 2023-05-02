package haproxy

import (
	"bufio"
	"crypto/tls"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/signalfx/golib/v3/datapoint"

	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// Map of HAProxy metrics name to their equivalent SignalFx names.
var sfxMetricsMap = map[string]string{
	"status":             haproxyStatus,
	"conn_tot":           haproxyConnectionTotal,
	"lbtot":              haproxyServerSelectedTotal,
	"bin":                haproxyBytesIn,
	"bout":               haproxyBytesOut,
	"cli_abrt":           haproxyClientAborts,
	"comp_byp":           haproxyCompressBypass,
	"comp_in":            haproxyCompressIn,
	"comp_out":           haproxyCompressOut,
	"comp_rsp":           haproxyCompressResponses,
	"CompressBpsIn":      haproxyCompressBitsPerSecondIn,
	"CompressBpsOut":     haproxyCompressBitsPerSecondOut,
	"CumConns":           haproxyConnections,
	"dreq":               haproxyDeniedRequest,
	"dresp":              haproxyDeniedResponse,
	"downtime":           haproxyDowntime,
	"econ":               haproxyErrorConnections,
	"ereq":               haproxyErrorRequest,
	"eresp":              haproxyErrorResponse,
	"chkfail":            haproxyFailedChecks,
	"wredis":             haproxyRedispatched,
	"req_tot":            haproxyRequestTotal,
	"CumReq":             haproxyRequests,
	"hrsp_1xx":           haproxyResponse1xx,
	"hrsp_2xx":           haproxyResponse2xx,
	"hrsp_3xx":           haproxyResponse3xx,
	"hrsp_4xx":           haproxyResponse4xx,
	"hrsp_5xx":           haproxyResponse5xx,
	"hrsp_other":         haproxyResponseOther,
	"wretr":              haproxyRetries,
	"stot":               haproxySessionTotal,
	"srv_abrt":           haproxyServerAborts,
	"SslCacheLookups":    haproxySslCacheLookups,
	"SslCacheMisses":     haproxySslCacheMisses,
	"CumSslConns":        haproxySslConnections,
	"Uptime_sec":         haproxyUptimeSeconds,
	"act":                haproxyActiveServers,
	"bck":                haproxyBackupServers,
	"check_duration":     haproxyCheckDuration,
	"conn_rate":          haproxyConnectionRate,
	"conn_rate_max":      haproxyConnectionRateMax,
	"CurrConns":          haproxyCurrentConnections,
	"CurrSslConns":       haproxyCurrentSslConnections,
	"dcon":               haproxyDeniedTCPConnections,
	"dses":               haproxyDeniedTCPSessions,
	"Idle_pct":           haproxyIdlePercent,
	"intercepted":        haproxyInterceptedRequests,
	"lastsess":           haproxyLastSession,
	"MaxConnRate":        haproxyMaxConnectionRate,
	"MaxConn":            haproxyMaxConnections,
	"MaxPipes":           haproxyMaxPipes,
	"MaxSessRate":        haproxyMaxSessionRate,
	"MaxSslConns":        haproxyMaxSslConnections,
	"PipesFree":          haproxyPipesFree,
	"PipesUsed":          haproxyPipesUsed,
	"qcur":               haproxyQueueCurrent,
	"qlimit":             haproxyQueueLimit,
	"qmax":               haproxyQueueMax,
	"qtime":              haproxyQueueTimeAverage,
	"req_rate":           haproxyRequestRate,
	"req_rate_max":       haproxyRequestRateMax,
	"rtime":              haproxyResponseTimeAverage,
	"Run_queue":          haproxyRunQueue,
	"scur":               haproxySessionCurrent,
	"slim":               haproxySessionLimit,
	"smax":               haproxySessionMax,
	"rate":               haproxySessionRate,
	"SessRate":           haproxySessionRateAll,
	"rate_lim":           haproxySessionRateLimit,
	"rate_max":           haproxySessionRateMax,
	"ttime":              haproxySessionTimeAverage,
	"SslBackendKeyRate":  haproxySslBackendKeyRate,
	"SslFrontendKeyRate": haproxySslFrontendKeyRate,
	"SslRate":            haproxySslRate,
	"Tasks":              haproxyTasks,
	"throttle":           haproxyThrottle,
	"ZlibMemUsage":       haproxyZlibMemoryUsage,
	"ConnRate":           haproxyConnectionRateAll,
}

// Fetches proxy stats datapoints from an http endpoint.
func (m *Monitor) statsHTTP(conf *Config, pxs proxies) []*datapoint.Datapoint {
	return m.statsHelper(conf, httpReader, "GET", pxs)
}

// Fetches proxy stats datapoints from a unix socket.
func (m *Monitor) statsSocket(conf *Config, pxs proxies) []*datapoint.Datapoint {
	return m.statsHelper(conf, socketReader, "show stat\n", pxs)
}

// A second order function for taking http and socket functions that fetch proxy stats datapoints.
func (m *Monitor) statsHelper(conf *Config, reader func(*Config, string) (io.ReadCloser, error), cmd string, proxies map[string]bool) []*datapoint.Datapoint {
	body, err := reader(conf, cmd)
	defer closeBody(body)
	if err != nil {
		m.logger.Error(err)
		return nil
	}
	dps := make([]*datapoint.Datapoint, 0)
	csvMap, err := statsMap(body)
	if err != nil {
		m.logger.Error(err)
		return nil
	}
	for _, headerValuePairs := range csvMap {
		if len(proxies) != 0 && !proxies[headerValuePairs["pxname"]] && !proxies[headerValuePairs["svname"]] {
			continue
		}
		for stat, value := range headerValuePairs {
			if dp := m.newDp(stat, value); dp != nil {
				dp.Dimensions["proxy_name"] = headerValuePairs["pxname"]
				dp.Dimensions["service_name"] = headerValuePairs["svname"]
				dp.Dimensions["type"] = headerValuePairs["type"]
				dp.Dimensions["server_id"] = headerValuePairs["sid"]
				dp.Dimensions["unique_proxy_id"] = headerValuePairs["iid"]
				// WARNING: Both pid and Process_num are HAProxy process identifiers. pid in the context of
				// proxy stats and Process_num in the context of HAProxy process info. It says in the docs
				// https://cbonte.github.io/haproxy-dconv/1.8/management.html#9.1 that pid is zero-based. But, we
				// find that pid is exactly the same as Process_num, a natural number. We therefore assign pid and
				// Process_num to dimension process_id without modifying them to match.
				dp.Dimensions["process_id"] = headerValuePairs["pid"]
				dps = append(dps, dp)
			}
		}
	}
	if len(dps) == 0 {
		m.logger.Errorf("failed to create proxy stats datapoints")
		return nil
	}
	return dps
}

// Fetches process info datapoints from a unix socket.
func (m *Monitor) infoSocket(conf *Config, _ proxies) []*datapoint.Datapoint {
	dps := make([]*datapoint.Datapoint, 0)
	infoPairs, err := infoMap(conf)
	if err != nil {
		m.logger.Error(err)
		return nil
	}
	for stat, value := range infoPairs {
		if dp := m.newDp(stat, value); dp != nil {
			// WARNING: Both pid and Process_num are HAProxy process identifiers. pid in the context of
			// proxy stats and Process_num in the context of HAProxy process info. It says in the docs
			// https://cbonte.github.io/haproxy-dconv/1.8/management.html#9.1 that pid is zero-based. But, we
			// find that pid is exactly the same as Process_num, a natural number. We therefore assign pid and
			// Process_num to dimension process_id without modifying them to match.
			dp.Dimensions["process_id"] = infoPairs["Process_num"]
			dps = append(dps, dp)
		}
	}
	return dps
}

// Fetches and convert proxy stats in csv format to map.
func statsMap(body io.Reader) (map[int]map[string]string, error) {
	r := csv.NewReader(body)
	r.TrimLeadingSpace = true
	r.TrailingComma = true
	rows := map[int]map[string]string{}
	table, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(table) == 0 {
		return nil, fmt.Errorf("unavailable proxy stats csv data")
	}
	if utils.TrimAllSpaces(table[0][0]) != "#pxname" {
		return nil, fmt.Errorf("incompatible proxy stats csv data. Expected '#pxname' as first header")
	}
	// fixing first column header because it is '# pxname' instead of 'pxname'
	table[0][0] = "pxname"
	for rowIndex := 1; rowIndex < len(table); rowIndex++ {
		if rows[rowIndex-1] == nil {
			rows[rowIndex-1] = map[string]string{}
		}
		for j, colName := range table[0] {
			rows[rowIndex-1][colName] = table[rowIndex][j]
		}
	}
	return rows, nil
}

// Fetches and converts process info (i.e. 'show info' command output) to map.
func infoMap(conf *Config) (map[string]string, error) {
	body, err := socketReader(conf, "show info\n")
	defer closeBody(body)
	if err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(body)
	processInfoOutput := map[string]string{}
	for sc.Scan() {
		s := strings.SplitN(sc.Text(), ":", 2)
		if len(s) != 2 || strings.TrimSpace(s[0]) == "" || strings.TrimSpace(s[1]) == "" {
			continue
		}
		processInfoOutput[strings.TrimSpace(s[0])] = strings.TrimSpace(s[1])
	}
	if len(processInfoOutput) == 0 {
		return nil, fmt.Errorf("failed to create process info datapoints")
	}
	return processInfoOutput, nil
}

// Creates datapoint from proxy stats and process info key value pairs.
func (m *Monitor) newDp(stat string, value string) *datapoint.Datapoint {
	metric := sfxMetricsMap[stat]
	if metric == "" || value == "" {
		return nil
	}
	dp := datapoint.New(metric, map[string]string{}, nil, metricSet[metric].Type, time.Time{})
	switch stat {
	case "status":
		dp.Value = datapoint.NewFloatValue(float64(parseStatusField(value)))
	default:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			switch err.(type) {
			case *strconv.NumError:
				m.logger.Debug(err)
			default:
				m.logger.Error(err)
			}
			return nil
		}
		dp.Value = datapoint.NewFloatValue(float64(v))
	}
	return dp
}

func parseStatusField(v string) int64 {
	switch v {
	case "UP", "UP 1/3", "UP 2/3", "OPEN", "no check":
		return 1
	case "DOWN", "DOWN 1/2", "NOLB", "MAINT":
		return 0
	}
	return 0
}

func httpReader(conf *Config, method string) (io.ReadCloser, error) {
	client := http.Client{
		Timeout:   conf.Timeout.AsDuration(),
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !conf.SSLVerify}},
	}
	req, err := http.NewRequest(method, conf.ScrapeURL(), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(conf.Username, conf.Password)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if !(res.StatusCode >= 200 && res.StatusCode < 300) {
		res.Body.Close()
		return nil, fmt.Errorf("HTTP status %d", res.StatusCode)
	}
	return res.Body, nil
}

func socketReader(conf *Config, cmd string) (io.ReadCloser, error) {
	u, err := url.Parse(conf.ScrapeURL())
	if err != nil {
		return nil, fmt.Errorf("cannot parse url %s status. %v", conf.ScrapeURL(), err)
	}
	f, err := net.DialTimeout("unix", u.Path, conf.Timeout.AsDuration())
	if err != nil {
		return nil, err
	}
	if err := f.SetDeadline(time.Now().Add(conf.Timeout.AsDuration())); err != nil {
		f.Close()
		return nil, err
	}
	n, err := io.WriteString(f, cmd)
	if err != nil {
		f.Close()
		return nil, err
	}
	if n != len(cmd) {
		f.Close()
		return nil, errors.New("write error")
	}
	return f, nil
}

func closeBody(body io.ReadCloser) {
	if body != nil {
		body.Close()
	}
}
