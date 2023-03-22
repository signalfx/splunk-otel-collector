package conviva

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	accountsURLFormat             = "https://api.conviva.com/insights/2.4/accounts.json"
	filtersURLFormat              = "https://api.conviva.com/insights/2.4/filters.json?account=%s"
	metricLensDimensionsURLFormat = "https://api.conviva.com/insights/2.4/metriclens_dimension_list.json?account=%s"
)

// account for Conviva account data
type account struct {
	id                   string
	Name                 string
	filters              map[string]string
	metricLensFilters    map[string]string
	metricLensDimensions map[string]float64
}

// accountsService interface for accounts related methods
type accountsService interface {
	getDefault() (*account, error)
	getMetricLensDimensionMap(accountName string) (map[string]float64, error)
	getID(accountName string) (string, error)
	getFilters(accountName string) (map[string]string, error)
	getMetricLensFilters(accountName string) (map[string]string, error)
	getFilterID(accountName string, filterName string) (string, error)
	getMetricLensDimensionID(accountName string, metricLensDimension string) (float64, error)
}

type accountsServiceImpl struct {
	defaultAccount      *account
	accounts            map[string]*account
	ctx                 context.Context
	timeout             *time.Duration
	client              httpClient
	mutex               sync.RWMutex
	accountsInitialized bool
}

// newAccountsService factory function creating accountsService
func newAccountsService(ctx context.Context, timeout *time.Duration, client httpClient) accountsService {
	service := accountsServiceImpl{ctx: ctx, timeout: timeout, client: client}
	return &service
}

func (s *accountsServiceImpl) getDefault() (*account, error) {
	if err := s.initAccounts(); err != nil {
		return nil, err
	}
	if s.defaultAccount == nil {
		return nil, fmt.Errorf("could not find default account")
	}
	return s.defaultAccount, nil
}

func (s *accountsServiceImpl) getMetricLensDimensionMap(accountName string) (map[string]float64, error) {
	if err := s.initAccounts(); err != nil {
		return nil, err
	}
	if a := s.accounts[accountName]; a != nil {
		return a.metricLensDimensions, nil
	}
	return nil, fmt.Errorf("could not find account %s", accountName)
}

func (s *accountsServiceImpl) getID(accountName string) (string, error) {
	if err := s.initAccounts(); err != nil {
		return "", err
	}
	if a := s.accounts[accountName]; a != nil {
		return a.id, nil
	}
	return "", fmt.Errorf("could not find account %s", accountName)
}

func (s *accountsServiceImpl) getFilters(accountName string) (map[string]string, error) {
	if err := s.initAccounts(); err != nil {
		return nil, err
	}
	if a := s.accounts[accountName]; a != nil {
		return a.filters, nil
	}
	return nil, fmt.Errorf("could not find account %s", accountName)
}

func (s *accountsServiceImpl) getMetricLensFilters(accountName string) (map[string]string, error) {
	if err := s.initAccounts(); err != nil {
		return nil, err
	}
	if a := s.accounts[accountName]; a != nil {
		return a.metricLensFilters, nil
	}
	return nil, fmt.Errorf("could not find account %s", accountName)
}

func (s *accountsServiceImpl) getFilterID(accountName string, filterName string) (string, error) {
	if err := s.initAccounts(); err != nil {
		return "", err
	}
	if a := s.accounts[accountName]; a != nil {
		for id, name := range a.filters {
			if name == filterName {
				return id, nil
			}
		}
	}
	return "", fmt.Errorf("no filter id for filter %s in account %s", filterName, accountName)
}

func (s *accountsServiceImpl) getMetricLensDimensionID(accountName string, metricLensDimension string) (float64, error) {
	if err := s.initAccounts(); err != nil {
		return 0, err
	}
	if a := s.accounts[accountName]; a != nil {
		if m := a.metricLensDimensions; m != nil {
			return m[metricLensDimension], nil
		}
	}
	return 0, fmt.Errorf("account %s has no MetricLens dimensions", accountName)
}

func (s *accountsServiceImpl) initAccounts() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !s.accountsInitialized {
		ctx, cancel := context.WithTimeout(s.ctx, *s.timeout)
		defer cancel()
		res := struct {
			Default  string            `json:"default"`
			Count    float64           `json:"count"`
			Accounts map[string]string `json:"accounts"`
		}{}
		if _, err := s.client.get(ctx, &res, accountsURLFormat); err != nil {
			return err
		}
		s.accounts = make(map[string]*account, len(res.Accounts))
		for name, id := range res.Accounts {
			a := account{Name: name, id: id}
			if a.Name == res.Default {
				s.defaultAccount = &a
			}
			s.accounts[name] = &a
		}
		if err := s.setAccountsFilters(); err != nil {
			return err
		}
		if err := s.setAccountsMetricLensDimensions(); err != nil {
			return err
		}
		if err := s.setAccountsMetricLensFilters(); err != nil {
			return err
		}
		s.accountsInitialized = true
	}
	return nil
}

func (s *accountsServiceImpl) setAccountsFilters() error {
	var g errgroup.Group
	for _, a := range s.accounts {
		newA := a
		newA.filters = map[string]string{}
		g.Go(func() error {
			ctx, cancel := context.WithTimeout(s.ctx, *s.timeout)
			defer cancel()
			if _, err := s.client.get(ctx, &newA.filters, fmt.Sprintf(filtersURLFormat, newA.id)); err != nil {
				return err
			}
			return nil
		})
	}
	return g.Wait()
}

func (s *accountsServiceImpl) setAccountsMetricLensDimensions() error {
	var g errgroup.Group
	for _, a := range s.accounts {
		newA := a
		newA.metricLensDimensions = map[string]float64{}
		g.Go(func() error {
			ctx, cancel := context.WithTimeout(s.ctx, *s.timeout)
			defer cancel()
			if _, err := s.client.get(ctx, &newA.metricLensDimensions, fmt.Sprintf(metricLensDimensionsURLFormat, newA.id)); err != nil {
				return err
			}
			return nil
		})
	}
	return g.Wait()
}

// setAccountsMetricLensFilters() creates a map of MetricLens enabled filters from the map of filters of accounts.
// A filter is MetricLens enabled if getting quality_metriclens metrics is successful (i.e. no errors and the status code is 200).
// A filter is not MetricLens enabled if the status code is 400. Otherwise an error occurred and that error gets returned.
func (s *accountsServiceImpl) setAccountsMetricLensFilters() error {
	var (
		g     errgroup.Group
		mutex sync.RWMutex
	)
	for _, a := range s.accounts {
		newA := a
		newA.metricLensFilters = map[string]string{}
		for filterID, filter := range newA.filters {
			newFilterID := filterID
			newFilter := filter
			g.Go(func() error {
				ctx, cancel := context.WithTimeout(s.ctx, *s.timeout)
				defer cancel()
				for _, dimID := range newA.metricLensDimensions {
					if statusCode, err := s.client.get(ctx, &map[string]interface{}{}, fmt.Sprintf(metricLensURLFormat, "quality_metriclens", newA.id, newFilterID, int(dimID))); err == nil && statusCode == 200 {
						mutex.Lock()
						newA.metricLensFilters[newFilterID] = newFilter
						mutex.Unlock()
					} else if statusCode < 400 || statusCode > 499 {
						return err
					}
					break
				}
				return nil
			})
		}
	}
	return g.Wait()
}
