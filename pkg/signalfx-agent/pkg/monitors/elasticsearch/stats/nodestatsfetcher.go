package stats

import (
	"github.com/signalfx/golib/v3/datapoint"
	esUtils "github.com/signalfx/signalfx-agent/pkg/monitors/elasticsearch/utils"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// Groups of Node stats being that the monitor collects
const (
	TransportStatsGroup  = "transport"
	HTTPStatsGroup       = "http"
	JVMStatsGroup        = "jvm"
	ThreadpoolStatsGroup = "thread_pool"
	ProcessStatsGroup    = "process"
)

// GetNodeStatsDatapoints fetches datapoints for ES Node stats
func GetNodeStatsDatapoints(nodeStatsOutput *NodeStatsOutput, defaultDims map[string]string, selectedThreadPools map[string]bool, enhancedStatsForIndexGroups map[string]bool, nodeStatsGroupEnhancedOption map[string]bool) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint
	for _, nodeStats := range nodeStatsOutput.NodeStats {
		out = append(out, getNodeStatsDatapointsHelper(nodeStats, defaultDims, selectedThreadPools, enhancedStatsForIndexGroups, nodeStatsGroupEnhancedOption)...)
	}
	return out
}

func getNodeStatsDatapointsHelper(nodeStats NodeStats, defaultDims map[string]string, selectedThreadPools map[string]bool, enhancedStatsForIndexGroups map[string]bool, nodeStatsGroupEnhancedOption map[string]bool) []*datapoint.Datapoint {
	var dps []*datapoint.Datapoint

	dps = append(dps, nodeStats.JVM.getJVMStats(nodeStatsGroupEnhancedOption[JVMStatsGroup], defaultDims)...)
	dps = append(dps, nodeStats.Process.getProcessStats(nodeStatsGroupEnhancedOption[ProcessStatsGroup], defaultDims)...)
	dps = append(dps, nodeStats.Transport.getTransportStats(nodeStatsGroupEnhancedOption[TransportStatsGroup], defaultDims)...)
	dps = append(dps, nodeStats.HTTP.getHTTPStats(nodeStatsGroupEnhancedOption[HTTPStatsGroup], defaultDims)...)
	dps = append(dps, fetchThreadPoolStats(nodeStatsGroupEnhancedOption[ThreadpoolStatsGroup], nodeStats.ThreadPool, defaultDims, selectedThreadPools)...)
	dps = append(dps, nodeStats.Indices.getIndexGroupStats(enhancedStatsForIndexGroups, defaultDims)...)

	return dps
}

func (jvm *JVM) getJVMStats(enhanced bool, dims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.threads.count", dims, jvm.JvmThreadsStats.Count),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.threads.peak", dims, jvm.JvmThreadsStats.PeakCount),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.heap-used-percent", dims, jvm.JvmMemStats.HeapUsedPercent),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.heap-max", dims, jvm.JvmMemStats.HeapMaxInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.non-heap-committed", dims, jvm.JvmMemStats.NonHeapCommittedInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.non-heap-used", dims, jvm.JvmMemStats.NonHeapUsedInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.pools.young.max_in_bytes", dims, jvm.JvmMemStats.Pools.Young.MaxInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.pools.young.used_in_bytes", dims, jvm.JvmMemStats.Pools.Young.UsedInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.pools.young.peak_used_in_bytes", dims, jvm.JvmMemStats.Pools.Young.PeakUsedInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.pools.young.peak_max_in_bytes", dims, jvm.JvmMemStats.Pools.Young.PeakMaxInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.pools.old.max_in_bytes", dims, jvm.JvmMemStats.Pools.Old.MaxInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.pools.old.used_in_bytes", dims, jvm.JvmMemStats.Pools.Old.UsedInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.pools.old.peak_used_in_bytes", dims, jvm.JvmMemStats.Pools.Old.PeakUsedInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.pools.old.peak_max_in_bytes", dims, jvm.JvmMemStats.Pools.Old.PeakMaxInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.pools.survivor.max_in_bytes", dims, jvm.JvmMemStats.Pools.Survivor.MaxInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.pools.survivor.used_in_bytes", dims, jvm.JvmMemStats.Pools.Survivor.UsedInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.pools.survivor.peak_used_in_bytes", dims, jvm.JvmMemStats.Pools.Survivor.PeakUsedInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.pools.survivor.peak_max_in_bytes", dims, jvm.JvmMemStats.Pools.Survivor.PeakMaxInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.buffer_pools.mapped.count", dims, jvm.BufferPools.Mapped.Count),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.buffer_pools.mapped.used_in_bytes", dims, jvm.BufferPools.Mapped.UsedInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.buffer_pools.mapped.total_capacity_in_bytes", dims, jvm.BufferPools.Mapped.TotalCapacityInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.buffer_pools.direct.count", dims, jvm.BufferPools.Direct.Count),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.buffer_pools.direct.used_in_bytes", dims, jvm.BufferPools.Direct.UsedInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.buffer_pools.direct.total_capacity_in_bytes", dims, jvm.BufferPools.Direct.TotalCapacityInBytes),
			esUtils.PrepareGaugeHelper("elasticsearch.jvm.classes.current-loaded-count", dims, jvm.Classes.CurrentLoadedCount),

			esUtils.PrepareCumulativeHelper("elasticsearch.jvm.gc.count", dims, jvm.JvmGcStats.Collectors.Young.CollectionCount),
			esUtils.PrepareCumulativeHelper("elasticsearch.jvm.gc.old-count", dims, jvm.JvmGcStats.Collectors.Old.CollectionCount),
			esUtils.PrepareCumulativeHelper("elasticsearch.jvm.gc.old-time", dims, jvm.JvmGcStats.Collectors.Old.CollectionTimeInMillis),
			esUtils.PrepareCumulativeHelper("elasticsearch.jvm.classes.total-loaded-count", dims, jvm.Classes.TotalLoadedCount),
			esUtils.PrepareCumulativeHelper("elasticsearch.jvm.classes.total-unloaded-count", dims, jvm.Classes.TotalUnloadedCount),
		}...)
	}

	out = append(out, []*datapoint.Datapoint{
		esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.heap-used", dims, jvm.JvmMemStats.HeapUsedInBytes),
		esUtils.PrepareGaugeHelper("elasticsearch.jvm.mem.heap-committed", dims, jvm.JvmMemStats.HeapCommittedInBytes),
		esUtils.PrepareCumulativeHelper("elasticsearch.jvm.uptime", dims, jvm.UptimeInMillis),
		esUtils.PrepareCumulativeHelper("elasticsearch.jvm.gc.time", dims, jvm.JvmGcStats.Collectors.Young.CollectionTimeInMillis),
	}...)

	return out
}

func (processStats *Process) getProcessStats(enhanced bool, dims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper("elasticsearch.process.max_file_descriptors", dims, processStats.MaxFileDescriptors),
			esUtils.PrepareGaugeHelper("elasticsearch.process.cpu.percent", dims, processStats.CPU.Percent),
			esUtils.PrepareCumulativeHelper("elasticsearch.process.cpu.time", dims, processStats.CPU.TotalInMillis),
			esUtils.PrepareCumulativeHelper("elasticsearch.process.mem.total-virtual-size", dims, processStats.Mem.TotalVirtualInBytes),
		}...)
	}

	out = append(out, []*datapoint.Datapoint{
		esUtils.PrepareGaugeHelper("elasticsearch.process.open_file_descriptors", dims, processStats.OpenFileDescriptors),
	}...)

	return out
}

func fetchThreadPoolStats(enhanced bool, threadPools map[string]ThreadPoolStats, defaultDims map[string]string, selectedThreadPools map[string]bool) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint
	for threadPool, stats := range threadPools {
		if !selectedThreadPools[threadPool] {
			continue
		}
		out = append(out, threadPoolDatapoints(enhanced, threadPool, stats, defaultDims)...)
	}
	return out
}

func threadPoolDatapoints(enhanced bool, threadPool string, threadPoolStats ThreadPoolStats, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint
	threadPoolDimension := map[string]string{}
	threadPoolDimension["thread_pool"] = threadPool

	dims := utils.MergeStringMaps(defaultDims, threadPoolDimension)

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper("elasticsearch.thread_pool.threads", dims, threadPoolStats.Threads),
			esUtils.PrepareGaugeHelper("elasticsearch.thread_pool.queue", dims, threadPoolStats.Queue),
			esUtils.PrepareGaugeHelper("elasticsearch.thread_pool.active", dims, threadPoolStats.Active),
			esUtils.PrepareGaugeHelper("elasticsearch.thread_pool.largest", dims, threadPoolStats.Largest),
			esUtils.PrepareCumulativeHelper("elasticsearch.thread_pool.completed", dims, threadPoolStats.Completed),
		}...)
	}

	out = append(out, []*datapoint.Datapoint{
		esUtils.PrepareCumulativeHelper("elasticsearch.thread_pool.rejected", dims, threadPoolStats.Rejected),
	}...)
	return out
}

func (transport *Transport) getTransportStats(enhanced bool, dims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper("elasticsearch.transport.server_open", dims, transport.ServerOpen),
			esUtils.PrepareCumulativeHelper("elasticsearch.transport.rx.count", dims, transport.RxCount),
			esUtils.PrepareCumulativeHelper("elasticsearch.transport.rx.size", dims, transport.RxSizeInBytes),
			esUtils.PrepareCumulativeHelper("elasticsearch.transport.tx.count", dims, transport.TxCount),
			esUtils.PrepareCumulativeHelper("elasticsearch.transport.tx.size", dims, transport.TxSizeInBytes),
		}...)
	}
	return out
}

func (http *HTTP) getHTTPStats(enhanced bool, dims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper("elasticsearch.http.current_open", dims, http.CurrentOpen),
			esUtils.PrepareCumulativeHelper("elasticsearch.http.total_open", dims, http.TotalOpened),
		}...)
	}

	return out
}

func (indexStatsGroups *IndexStatsGroups) getIndexGroupStats(enhancedStatsForIndexGroups map[string]bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	out = append(out, indexStatsGroups.Docs.getDocsStats(enhancedStatsForIndexGroups[DocsStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Store.getStoreStats(enhancedStatsForIndexGroups[StoreStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Indexing.getIndexingStats(enhancedStatsForIndexGroups[IndexingStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Get.getGetStats(enhancedStatsForIndexGroups[GetStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Search.getSearchStats(enhancedStatsForIndexGroups[SearchStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Merges.getMergesStats(enhancedStatsForIndexGroups[MergesStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Refresh.getRefreshStats(enhancedStatsForIndexGroups[RefreshStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Flush.getFlushStats(enhancedStatsForIndexGroups[FlushStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Warmer.getWarmerStats(enhancedStatsForIndexGroups[WarmerStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.QueryCache.getQueryCacheStats(enhancedStatsForIndexGroups[QueryCacheStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.FilterCache.getFilterCacheStats(enhancedStatsForIndexGroups[FilterCacheStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Fielddata.getFielddataStats(enhancedStatsForIndexGroups[FieldDataStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Completion.getCompletionStats(enhancedStatsForIndexGroups[CompletionStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Segments.getSegmentsStats(enhancedStatsForIndexGroups[SegmentsStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Translog.getTranslogStats(enhancedStatsForIndexGroups[TranslogStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.RequestCache.getRequestCacheStats(enhancedStatsForIndexGroups[RequestCacheStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Recovery.getRecoveryStats(enhancedStatsForIndexGroups[RecoveryStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.IDCache.getIDCacheStats(enhancedStatsForIndexGroups[IDCacheStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Suggest.getSuggestStats(enhancedStatsForIndexGroups[SuggestStatsGroup], defaultDims)...)
	out = append(out, indexStatsGroups.Percolate.getPercolateStats(enhancedStatsForIndexGroups[PercolateStatsGroup], defaultDims)...)

	return out
}

func (docs *Docs) getDocsStats(_ bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	out = append(out, []*datapoint.Datapoint{
		esUtils.PrepareGaugeHelper(elasticsearchIndicesDocsCount, defaultDims, docs.Count),
		esUtils.PrepareGaugeHelper(elasticsearchIndicesDocsDeleted, defaultDims, docs.Deleted),
	}...)

	return out
}

func (store *Store) getStoreStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesStoreThrottleTime, defaultDims, store.ThrottleTimeInMillis),
		}...)
	}

	out = append(out, []*datapoint.Datapoint{
		esUtils.PrepareGaugeHelper(elasticsearchIndicesStoreSize, defaultDims, store.SizeInBytes),
	}...)

	return out
}

func (indexing Indexing) getIndexingStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesIndexingIndexCurrent, defaultDims, indexing.IndexCurrent),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesIndexingIndexFailed, defaultDims, indexing.IndexFailed),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesIndexingDeleteCurrent, defaultDims, indexing.DeleteCurrent),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesIndexingDeleteTotal, defaultDims, indexing.DeleteTotal),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesIndexingDeleteTime, defaultDims, indexing.DeleteTimeInMillis),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesIndexingNoopUpdateTotal, defaultDims, indexing.NoopUpdateTotal),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesIndexingThrottleTime, defaultDims, indexing.ThrottleTimeInMillis),
		}...)
	}

	out = append(out, []*datapoint.Datapoint{
		esUtils.PrepareCumulativeHelper(elasticsearchIndicesIndexingIndexTotal, defaultDims, indexing.IndexTotal),
		esUtils.PrepareCumulativeHelper(elasticsearchIndicesIndexingIndexTime, defaultDims, indexing.IndexTimeInMillis),
	}...)

	return out
}

func (get *Get) getGetStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesGetCurrent, defaultDims, get.Current),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesGetTime, defaultDims, get.TimeInMillis),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesGetExistsTotal, defaultDims, get.ExistsTotal),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesGetExistsTime, defaultDims, get.ExistsTimeInMillis),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesGetMissingTotal, defaultDims, get.MissingTotal),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesGetMissingTime, defaultDims, get.MissingTimeInMillis),
		}...)
	}

	out = append(out, []*datapoint.Datapoint{
		esUtils.PrepareCumulativeHelper(elasticsearchIndicesGetTotal, defaultDims, get.Total),
	}...)

	return out
}

func (search *Search) getSearchStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSearchQueryCurrent, defaultDims, search.QueryCurrent),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSearchFetchCurrent, defaultDims, search.FetchCurrent),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSearchScrollCurrent, defaultDims, search.ScrollCurrent),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSearchSuggestCurrent, defaultDims, search.SuggestCurrent),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSearchOpenContexts, defaultDims, search.SuggestCurrent),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesSearchFetchTime, defaultDims, search.FetchTimeInMillis),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesSearchFetchTotal, defaultDims, search.FetchTotal),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesSearchScrollTime, defaultDims, search.ScrollTimeInMillis),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesSearchScrollTotal, defaultDims, search.ScrollTotal),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesSearchSuggestTime, defaultDims, search.SuggestTimeInMillis),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesSearchSuggestTotal, defaultDims, search.SuggestTotal),
		}...)
	}

	out = append(out, []*datapoint.Datapoint{
		esUtils.PrepareCumulativeHelper(elasticsearchIndicesSearchQueryTime, defaultDims, search.QueryTimeInMillis),
		esUtils.PrepareCumulativeHelper(elasticsearchIndicesSearchQueryTotal, defaultDims, search.QueryTotal),
	}...)

	return out
}

func (merges *Merges) getMergesStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesMergesCurrentDocs, defaultDims, merges.CurrentDocs),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesMergesCurrentSize, defaultDims, merges.CurrentSizeInBytes),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesMergesTotalDocs, defaultDims, merges.TotalDocs),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesMergesTotalSize, defaultDims, merges.TotalSizeInBytes),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesMergesStoppedTime, defaultDims, merges.TotalStoppedTimeInMillis),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesMergesThrottleTime, defaultDims, merges.TotalThrottledTimeInMillis),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesMergesAutoThrottleSize, defaultDims, merges.TotalAutoThrottleInBytes),
		}...)
	}

	out = append(out, []*datapoint.Datapoint{
		esUtils.PrepareGaugeHelper(elasticsearchIndicesMergesCurrent, defaultDims, merges.Current),
		esUtils.PrepareCumulativeHelper(elasticsearchIndicesMergesTotal, defaultDims, merges.Total),
		esUtils.PrepareCumulativeHelper(elasticsearchIndicesMergesTotalTime, defaultDims, merges.TotalTimeInMillis),
	}...)

	return out
}

func (refresh *Refresh) getRefreshStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesRefreshListeners, defaultDims, refresh.Listeners),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesRefreshTotal, defaultDims, refresh.Total),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesRefreshTotalTime, defaultDims, refresh.TotalTimeInMillis),
		}...)
	}

	return out
}

func (flush *Flush) getFlushStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesFlushPeriodic, defaultDims, flush.Periodic),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesFlushTotal, defaultDims, flush.Total),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesFlushTotalTime, defaultDims, flush.TotalTimeInMillis),
		}...)
	}

	return out
}

func (warmer *Warmer) getWarmerStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesWarmerCurrent, defaultDims, warmer.Current),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesWarmerTotal, defaultDims, warmer.Total),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesWarmerTotalTime, defaultDims, warmer.TotalTimeInMillis),
		}...)
	}

	return out
}

func (queryCache *QueryCache) getQueryCacheStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesQueryCacheCacheSize, defaultDims, queryCache.CacheSize),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesQueryCacheCacheCount, defaultDims, queryCache.CacheCount),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesQueryCacheEvictions, defaultDims, queryCache.Evictions),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesQueryCacheHitCount, defaultDims, queryCache.HitCount),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesQueryCacheMissCount, defaultDims, queryCache.MissCount),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesQueryCacheTotalCount, defaultDims, queryCache.TotalCount),
		}...)
	}
	out = append(out, []*datapoint.Datapoint{
		esUtils.PrepareGaugeHelper(elasticsearchIndicesQueryCacheMemorySize, defaultDims, queryCache.MemorySizeInBytes),
	}...)

	return out
}

func (filterCache *FilterCache) getFilterCacheStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesFilterCacheEvictions, defaultDims, filterCache.Evictions),
		}...)

		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesFilterCacheMemorySize, defaultDims, filterCache.MemorySizeInBytes),
		}...)
	}

	return out
}

func (fielddata *Fielddata) getFielddataStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesFielddataEvictions, defaultDims, fielddata.Evictions),
		}...)
	}

	out = append(out, []*datapoint.Datapoint{
		esUtils.PrepareGaugeHelper(elasticsearchIndicesFielddataMemorySize, defaultDims, fielddata.MemorySizeInBytes),
	}...)

	return out
}

func (completion *Completion) getCompletionStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesCompletionSize, defaultDims, completion.SizeInBytes),
		}...)
	}

	return out
}

func (segments *Segments) getSegmentsStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSegmentsMemorySize, defaultDims, segments.MemoryInBytes),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSegmentsIndexWriterMemorySize, defaultDims, segments.IndexWriterMemoryInBytes),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSegmentsIndexWriterMaxMemorySize, defaultDims, segments.IndexWriterMaxMemoryInBytes),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSegmentsVersionMapMemorySize, defaultDims, segments.VersionMapMemoryInBytes),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSegmentsTermsMemorySize, defaultDims, segments.TermsMemoryInBytes),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSegmentsStoredFieldMemorySize, defaultDims, segments.StoredFieldsMemoryInBytes),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSegmentsTermVectorsMemorySize, defaultDims, segments.TermVectorsMemoryInBytes),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSegmentsNormsMemorySize, defaultDims, segments.NormsMemoryInBytes),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSegmentsPointsMemorySize, defaultDims, segments.PointsMemoryInBytes),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSegmentsDocValuesMemorySize, defaultDims, segments.DocValuesMemoryInBytes),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSegmentsFixedBitSetMemorySize, defaultDims, segments.FixedBitSetMemoryInBytes),
		}...)
	}

	out = append(out, []*datapoint.Datapoint{
		esUtils.PrepareGaugeHelper(elasticsearchIndicesSegmentsCount, defaultDims, segments.Count),
	}...)

	return out
}

func (translog *Translog) getTranslogStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesTranslogUncommittedOperations, defaultDims, translog.UncommittedOperations),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesTranslogUncommittedSizeInBytes, defaultDims, translog.UncommittedSizeInBytes),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesTranslogEarliestLastModifiedAge, defaultDims, translog.EarliestLastModifiedAge),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesTranslogOperations, defaultDims, translog.Operations),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesTranslogSize, defaultDims, translog.SizeInBytes),
		}...)
	}

	return out
}

func (requestCache *RequestCache) getRequestCacheStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesRequestCacheEvictions, defaultDims, requestCache.Evictions),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesRequestCacheHitCount, defaultDims, requestCache.HitCount),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesRequestCacheMissCount, defaultDims, requestCache.MissCount),
		}...)
	}

	out = append(out, esUtils.PrepareGaugeHelper(elasticsearchIndicesRequestCacheMemorySize, defaultDims, requestCache.MemorySizeInBytes))

	return out
}

func (recovery *Recovery) getRecoveryStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesRecoveryCurrentAsSource, defaultDims, recovery.CurrentAsSource),
			esUtils.PrepareGaugeHelper(elasticsearchIndicesRecoveryCurrentAsTarget, defaultDims, recovery.CurrentAsTarget),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesRecoveryThrottleTime, defaultDims, recovery.ThrottleTimeInMillis),
		}...)
	}

	return out
}

func (idCache *IDCache) getIDCacheStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesIDCacheMemorySize, defaultDims, idCache.MemorySizeInBytes),
		}...)
	}

	return out
}

func (suggest *Suggest) getSuggestStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesSuggestCurrent, defaultDims, suggest.Current),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesSuggestTime, defaultDims, suggest.TimeInMillis),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesSuggestTotal, defaultDims, suggest.Total),
		}...)
	}

	return out
}

func (percolate *Percolate) getPercolateStats(enhanced bool, defaultDims map[string]string) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enhanced {
		out = append(out, []*datapoint.Datapoint{
			esUtils.PrepareGaugeHelper(elasticsearchIndicesPercolateCurrent, defaultDims, percolate.Current),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesPercolateTotal, defaultDims, percolate.Total),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesPercolateQueries, defaultDims, percolate.Queries),
			esUtils.PrepareCumulativeHelper(elasticsearchIndicesPercolateTime, defaultDims, percolate.TimeInMillis),
		}...)
	}

	return out
}
