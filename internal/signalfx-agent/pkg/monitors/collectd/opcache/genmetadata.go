// Code generated by monitor-code-gen. DO NOT EDIT.

package opcache

import (
	"github.com/signalfx/golib/v3/datapoint"

	"github.com/signalfx/signalfx-agent/pkg/monitors"
)

const monitorType = "collectd/opcache"

var groupSet = map[string]bool{}

const (
	cacheRatioOpcacheStatisticsOpcacheHitRate = "cache_ratio.opcache_statistics-opcache_hit_rate"
	cacheResultOpcacheStatisticsHits          = "cache_result.opcache_statistics-hits"
	cacheResultOpcacheStatisticsMisses        = "cache_result.opcache_statistics-misses"
	cacheSizeMemoryUsageFreeMemory            = "cache_size.memory_usage-free_memory"
	cacheSizeMemoryUsageUsedMemory            = "cache_size.memory_usage-used_memory"
	filesOpcacheStatisticsMaxCachedKeys       = "files.opcache_statistics-max_cached_keys"
	filesOpcacheStatisticsNumCachedKeys       = "files.opcache_statistics-num_cached_keys"
	filesOpcacheStatisticsNumCachedScripts    = "files.opcache_statistics-num_cached_scripts"
)

var metricSet = map[string]monitors.MetricInfo{
	cacheRatioOpcacheStatisticsOpcacheHitRate: {Type: datapoint.Gauge},
	cacheResultOpcacheStatisticsHits:          {Type: datapoint.Counter},
	cacheResultOpcacheStatisticsMisses:        {Type: datapoint.Counter},
	cacheSizeMemoryUsageFreeMemory:            {Type: datapoint.Gauge},
	cacheSizeMemoryUsageUsedMemory:            {Type: datapoint.Gauge},
	filesOpcacheStatisticsMaxCachedKeys:       {Type: datapoint.Gauge},
	filesOpcacheStatisticsNumCachedKeys:       {Type: datapoint.Gauge},
	filesOpcacheStatisticsNumCachedScripts:    {Type: datapoint.Gauge},
}

var defaultMetrics = map[string]bool{
	cacheRatioOpcacheStatisticsOpcacheHitRate: true,
	cacheResultOpcacheStatisticsHits:          true,
	cacheResultOpcacheStatisticsMisses:        true,
	cacheSizeMemoryUsageFreeMemory:            true,
	cacheSizeMemoryUsageUsedMemory:            true,
	filesOpcacheStatisticsMaxCachedKeys:       true,
	filesOpcacheStatisticsNumCachedKeys:       true,
	filesOpcacheStatisticsNumCachedScripts:    true,
}

var groupMetricsMap = map[string][]string{}

var monitorMetadata = monitors.Metadata{
	MonitorType:     "collectd/opcache",
	DefaultMetrics:  defaultMetrics,
	Metrics:         metricSet,
	SendUnknown:     false,
	Groups:          groupSet,
	GroupMetricsMap: groupMetricsMap,
	SendAll:         false,
}
