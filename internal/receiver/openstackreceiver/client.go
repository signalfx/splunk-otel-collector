// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openstackreceiver

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/hypervisors"
	novaLimits "github.com/gophercloud/gophercloud/v2/openstack/compute/v2/limits"
	"github.com/gophercloud/gophercloud/v2/openstack/identity/v3/tokens"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/groups"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
)

// openstackClient abstracts OpenStack API calls, enabling mock injection in tests.
type openstackClient interface {
	// Nova
	GetNovaLimits(ctx context.Context, projectID string) (*novaLimits.Limits, error)
	ListHypervisors(ctx context.Context) ([]hypervisors.Hypervisor, error)

	// Neutron
	CountNetworks(ctx context.Context) (int, error)
	CountSubnets(ctx context.Context) (int, error)
	CountRouters(ctx context.Context) (int, error)
	CountFloatingIPs(ctx context.Context) (int, error)
	CountSecurityGroups(ctx context.Context) (int, error)

	// Auth context
	ProjectID() string
	ProjectName() string
	ProjectDomainName() string
	UserDomainName() string
}

type gophercloudClient struct {
	novaClient    *gophercloud.ServiceClient
	neutronClient *gophercloud.ServiceClient

	projectID         string
	projectName       string
	projectDomainName string
	userDomainName    string
}

func newOpenstackClient(ctx context.Context, cfg *Config) (openstackClient, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.TLSInsecureSkipVerify, //nolint:gosec
	}
	provider, err := openstack.NewClient(cfg.AuthURL)
	if err != nil {
		return nil, err
	}
	provider.HTTPClient = http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   cfg.HTTPTimeout,
	}

	authOpts := gophercloud.AuthOptions{
		IdentityEndpoint: cfg.AuthURL,
		Username:         cfg.Username,
		Password:         string(cfg.Password),
		DomainID:         cfg.UserDomainID,
		Scope: &gophercloud.AuthScope{
			ProjectName: cfg.ProjectName,
			DomainID:    cfg.ProjectDomainID,
		},
	}

	if authErr := openstack.Authenticate(ctx, provider, authOpts); authErr != nil {
		return nil, authErr
	}

	endpointOpts := gophercloud.EndpointOpts{
		Region: cfg.RegionName,
	}

	novaClient, err := openstack.NewComputeV2(provider, endpointOpts)
	if err != nil {
		return nil, err
	}

	neutronClient, err := openstack.NewNetworkV2(provider, endpointOpts)
	if err != nil {
		return nil, err
	}

	// Use configured values as attribute defaults (matching collectd/openstack behavior).
	// project_id is filled from the token where available.
	projectID := ""
	projectName := cfg.ProjectName
	projectDomainName := cfg.ProjectDomainID
	userDomainName := cfg.UserDomainID

	// Resolve the actual project ID (and optionally domain names) from the token.
	identityClient, identErr := openstack.NewIdentityV3(provider, endpointOpts)
	if identErr != nil {
		return nil, identErr
	}
	tokenResult := tokens.Get(ctx, identityClient, provider.Token())
	if project, tokErr := tokenResult.ExtractProject(); tokErr == nil && project != nil {
		projectID = project.ID
		if project.Name != "" {
			projectName = project.Name
		}
		if project.Domain.Name != "" {
			projectDomainName = project.Domain.Name
		}
	}
	if user, tokErr := tokenResult.ExtractUser(); tokErr == nil && user != nil {
		if user.Domain.Name != "" {
			userDomainName = user.Domain.Name
		}
	}

	return &gophercloudClient{
		novaClient:        novaClient,
		neutronClient:     neutronClient,
		projectID:         projectID,
		projectName:       projectName,
		projectDomainName: projectDomainName,
		userDomainName:    userDomainName,
	}, nil
}

func (c *gophercloudClient) ProjectID() string         { return c.projectID }
func (c *gophercloudClient) ProjectName() string       { return c.projectName }
func (c *gophercloudClient) ProjectDomainName() string { return c.projectDomainName }
func (c *gophercloudClient) UserDomainName() string    { return c.userDomainName }

func (c *gophercloudClient) GetNovaLimits(ctx context.Context, projectID string) (*novaLimits.Limits, error) {
	opts := novaLimits.GetOpts{}
	if projectID != "" {
		opts.TenantID = projectID
	}
	return novaLimits.Get(ctx, c.novaClient, opts).Extract()
}

func (c *gophercloudClient) ListHypervisors(ctx context.Context) ([]hypervisors.Hypervisor, error) {
	allPages, err := hypervisors.List(c.novaClient, hypervisors.ListOpts{}).AllPages(ctx)
	if err != nil {
		return nil, err
	}
	return hypervisors.ExtractHypervisors(allPages)
}

func (c *gophercloudClient) CountNetworks(ctx context.Context) (int, error) {
	allPages, err := networks.List(c.neutronClient, networks.ListOpts{}).AllPages(ctx)
	if err != nil {
		return 0, err
	}
	all, err := networks.ExtractNetworks(allPages)
	return len(all), err
}

func (c *gophercloudClient) CountSubnets(ctx context.Context) (int, error) {
	allPages, err := subnets.List(c.neutronClient, subnets.ListOpts{}).AllPages(ctx)
	if err != nil {
		return 0, err
	}
	all, err := subnets.ExtractSubnets(allPages)
	return len(all), err
}

func (c *gophercloudClient) CountRouters(ctx context.Context) (int, error) {
	allPages, err := routers.List(c.neutronClient, routers.ListOpts{}).AllPages(ctx)
	if err != nil {
		return 0, err
	}
	all, err := routers.ExtractRouters(allPages)
	return len(all), err
}

func (c *gophercloudClient) CountFloatingIPs(ctx context.Context) (int, error) {
	allPages, err := floatingips.List(c.neutronClient, floatingips.ListOpts{}).AllPages(ctx)
	if err != nil {
		return 0, err
	}
	all, err := floatingips.ExtractFloatingIPs(allPages)
	return len(all), err
}

func (c *gophercloudClient) CountSecurityGroups(ctx context.Context) (int, error) {
	allPages, err := groups.List(c.neutronClient, groups.ListOpts{}).AllPages(ctx)
	if err != nil {
		return 0, err
	}
	all, err := groups.ExtractGroups(allPages)
	return len(all), err
}
