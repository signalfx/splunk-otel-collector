package conviva

import (
	"fmt"
	"strings"
	"sync"
)

func metricLensMetrics() map[string][]string {
	return map[string][]string{
		"quality_metriclens":  groupMetricsMap[groupQualityMetriclens],
		"audience_metriclens": groupMetricsMap[groupAudienceMetriclens],
	}
}

// metricConfig for configuring individual metric
type metricConfig struct {
	// Conviva customer account name. The default account is fetched used if not specified.
	Account         string `yaml:"account"`
	MetricParameter string `yaml:"metricParameter" default:"quality_metriclens"`
	// Filter names. The default is `All Traffic` filter
	Filters []string `yaml:"filters"`
	// MetricLens dimension names. The default is names of all MetricLens dimensions of the account
	MetricLensDimensions []string `yaml:"metricLensDimensions"`
	// MetricLens dimension names to exclude.
	ExcludeMetricLensDimensions []string `yaml:"excludeMetricLensDimensions"`
	// Max number of filters per request. The default is the number of filters. Multiple requests are made if the number of filters is more than maxFiltersPerRequest
	MaxFiltersPerRequest int `yaml:"maxFiltersPerRequest"`
	accountID            string
	// id:name map of filters derived from the configured Filters
	filtersMap map[string]string
	// name:id map of MetricLens dimensions derived from configured MetricLensDimensions
	metricLensDimensionMap map[string]float64
	isInitialized          bool
	mutex                  sync.RWMutex
}

func (mc *metricConfig) filterName(filterID string) string {
	if len(mc.filtersMap) != 0 {
		return mc.filtersMap[filterID]
	}
	return ""
}

func (mc *metricConfig) init(service accountsService) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	if !mc.isInitialized {
		if err := mc.setAccount(service); err != nil {
			return fmt.Errorf("metric %s account setting failure. %+v", mc.MetricParameter, err)
		}
		if err := mc.setFilters(service); err != nil {
			return fmt.Errorf("metric %s filter(s) setting failure. %+v", mc.MetricParameter, err)
		}
		if err := mc.setMetricLensDimensions(service); err != nil {
			return fmt.Errorf("metric %s MetricLens dimension(s) setting failure. %+v", mc.MetricParameter, err)
		}
		if err := mc.excludeMetricLensDimensions(service); err != nil {
			return fmt.Errorf("metric %s MetricLens dimension(s) exclusion failure. %+v", mc.MetricParameter, err)
		}
		mc.isInitialized = true
	}
	return nil
}

// setting account id and default account if necessary
func (mc *metricConfig) setAccount(service accountsService) error {
	mc.Account = strings.TrimSpace(mc.Account)
	if mc.Account == "" {
		if defaultAccount, err := service.getDefault(); err == nil {
			mc.Account = defaultAccount.Name
		} else {
			return err
		}
	}
	var err error
	if mc.accountID, err = service.getID(mc.Account); err != nil {
		return err
	}
	return nil
}

func (mc *metricConfig) setFilters(service accountsService) error {
	switch {
	case len(mc.Filters) == 0:
		mc.Filters = []string{"All Traffic"}
		if id, err := service.getFilterID(mc.Account, "All Traffic"); err == nil {
			mc.filtersMap = map[string]string{id: "All Traffic"}
		} else {
			return err
		}
	case strings.TrimSpace(mc.Filters[0]) == "_ALL_":
		var allFilters map[string]string
		var err error
		if strings.Contains(mc.MetricParameter, "metriclens") {
			if allFilters, err = service.getMetricLensFilters(mc.Account); err != nil {
				return err
			}
		} else {
			if allFilters, err = service.getFilters(mc.Account); err != nil {
				return err
			}
		}
		mc.Filters = make([]string, 0, len(allFilters))
		mc.filtersMap = make(map[string]string, len(allFilters))
		for id, name := range allFilters {
			mc.Filters = append(mc.Filters, name)
			mc.filtersMap[id] = name
		}
	default:
		mc.filtersMap = make(map[string]string, len(mc.Filters))
		for _, name := range mc.Filters {
			name = strings.TrimSpace(name)
			if id, err := service.getFilterID(mc.Account, name); err == nil {
				mc.filtersMap[id] = name
			} else {
				return err
			}
		}
	}
	return nil
}

func (mc *metricConfig) setMetricLensDimensions(service accountsService) error {
	if strings.Contains(mc.MetricParameter, "metriclens") {
		if len(mc.MetricLensDimensions) == 0 || strings.TrimSpace(mc.MetricLensDimensions[0]) == "_ALL_" {
			if metricLensDimensionMap, err := service.getMetricLensDimensionMap(mc.Account); err == nil {
				mc.MetricLensDimensions = make([]string, 0, len(metricLensDimensionMap))
				mc.metricLensDimensionMap = make(map[string]float64, len(metricLensDimensionMap))
				for name, id := range metricLensDimensionMap {
					mc.MetricLensDimensions = append(mc.MetricLensDimensions, name)
					mc.metricLensDimensionMap[name] = id
				}
			} else {
				return err
			}

		} else {
			mc.metricLensDimensionMap = make(map[string]float64, len(mc.MetricLensDimensions))
			for i, name := range mc.MetricLensDimensions {
				name := strings.TrimSpace(name)
				if id, err := service.getMetricLensDimensionID(mc.Account, name); err == nil {
					mc.MetricLensDimensions[i] = name
					mc.metricLensDimensionMap[name] = id
				} else {
					return err
				}
			}
		}
	}
	return nil
}

func (mc *metricConfig) excludeMetricLensDimensions(service accountsService) error {
	for _, excludeName := range mc.ExcludeMetricLensDimensions {
		excludeName := strings.TrimSpace(excludeName)
		if _, err := service.getMetricLensDimensionID(mc.Account, excludeName); err == nil {
			delete(mc.metricLensDimensionMap, excludeName)
		} else {
			return err
		}
	}
	if len(mc.metricLensDimensionMap) < len(mc.MetricLensDimensions) {
		mc.MetricLensDimensions = make([]string, 0, len(mc.metricLensDimensionMap))
		for name := range mc.metricLensDimensionMap {
			mc.MetricLensDimensions = append(mc.MetricLensDimensions, name)
		}
	}
	return nil
}

func (mc *metricConfig) filterIDs() []string {
	ids := make([]string, 0, len(mc.filtersMap))
	for id := range mc.filtersMap {
		ids = append(ids, id)
	}
	return ids
}
