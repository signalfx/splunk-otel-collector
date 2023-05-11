package stats

import (
	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/signalfx-agent/pkg/utils"
)

// Valid Index stats groups
const (
	DocsStatsGroup         = "docs"
	StoreStatsGroup        = "store"
	IndexingStatsGroup     = "indexing"
	GetStatsGroup          = "get"
	SearchStatsGroup       = "search"
	MergesStatsGroup       = "merges"
	RefreshStatsGroup      = "refresh"
	FlushStatsGroup        = "flush"
	WarmerStatsGroup       = "warmer"
	QueryCacheStatsGroup   = "query_cache"
	FilterCacheStatsGroup  = "filter_cache"
	FieldDataStatsGroup    = "fielddata"
	CompletionStatsGroup   = "completion"
	SegmentsStatsGroup     = "segments"
	TranslogStatsGroup     = "translog"
	RequestCacheStatsGroup = "request_cache"
	RecoveryStatsGroup     = "recovery"
	IDCacheStatsGroup      = "id_cache"
	SuggestStatsGroup      = "suggest"
	PercolateStatsGroup    = "percolate"
)

// ValidIndexStatsGroups is a "set" of valid index stats groups
var ValidIndexStatsGroups = map[string]bool{
	DocsStatsGroup:         true,
	StoreStatsGroup:        true,
	IndexingStatsGroup:     true,
	GetStatsGroup:          true,
	SearchStatsGroup:       true,
	MergesStatsGroup:       true,
	RefreshStatsGroup:      true,
	FlushStatsGroup:        true,
	WarmerStatsGroup:       true,
	QueryCacheStatsGroup:   true,
	FilterCacheStatsGroup:  true,
	FieldDataStatsGroup:    true,
	CompletionStatsGroup:   true,
	SegmentsStatsGroup:     true,
	TranslogStatsGroup:     true,
	RequestCacheStatsGroup: true,
	RecoveryStatsGroup:     true,
	IDCacheStatsGroup:      true,
	SuggestStatsGroup:      true,
	PercolateStatsGroup:    true,
}

// Aggregations types for index stats
const (
	Total     = "total"
	Primaries = "primaries"
)

// GetIndexStatsSummaryDatapoints fetches datapoints for ES Index stats summary aggregated across all indexes
func GetIndexStatsSummaryDatapoints(indexStats IndexStats, defaultDims map[string]string, enhancedStatsForIndexGroupsOption map[string]bool, enablePrimaryIndexStats bool) []*datapoint.Datapoint {
	return getIndexStatsHelper(indexStats, defaultDims, enhancedStatsForIndexGroupsOption, enablePrimaryIndexStats)
}

// GetIndexStatsDatapoints fetches datapoints for ES Index stats per index
func GetIndexStatsDatapoints(indexStatsPerIndex map[string]IndexStats, indexes map[string]bool, defaultDims map[string]string, enhancedStatsForIndexGroupsOption map[string]bool, enablePrimaryIndexStats bool) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint
	collectAllIndexes := len(indexes) == 0

	for indexName, indexStats := range indexStatsPerIndex {
		if !collectAllIndexes && !indexes[indexName] {
			continue
		}

		dims := utils.MergeStringMaps(defaultDims, map[string]string{
			"index": indexName,
		})
		out = append(out, getIndexStatsHelper(indexStats, dims, enhancedStatsForIndexGroupsOption, enablePrimaryIndexStats)...)
	}

	return out
}

func getIndexStatsHelper(indexStats IndexStats, defaultDims map[string]string, enhancedStatsForIndexGroupsOption map[string]bool, enablePrimaryIndexStats bool) []*datapoint.Datapoint {
	var out []*datapoint.Datapoint

	if enablePrimaryIndexStats {
		indexStatsGroup := indexStats.Primaries
		defaultDimsForPrimaries := utils.MergeStringMaps(defaultDims, map[string]string{
			"aggregation": Primaries,
		})
		out = append(out, indexStatsGroup.getIndexGroupStats(enhancedStatsForIndexGroupsOption, defaultDimsForPrimaries)...)
	}

	indexStatsTotalGroup := indexStats.Total
	defaultDimsForTotal := utils.MergeStringMaps(defaultDims, map[string]string{
		"aggregation": Total,
	})
	out = append(out, indexStatsTotalGroup.getIndexGroupStats(enhancedStatsForIndexGroupsOption, defaultDimsForTotal)...)

	return out
}
