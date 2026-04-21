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
	"errors"
	"fmt"
	"maps"
	"strconv"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

var errInvalidConfig = errors.New("invalid config: expected *Config")

// newClientFn allows tests to inject a mock client.
var newClientFn = newOpenstackClient

type openstackScraper struct {
	logger *zap.Logger
	cfg    *Config
	client openstackClient
}

func newScraper(logger *zap.Logger, cfg *Config) *openstackScraper {
	return &openstackScraper{logger: logger, cfg: cfg}
}

func (s *openstackScraper) start(ctx context.Context, _ component.Host) error {
	client, err := newClientFn(ctx, s.cfg)
	if err != nil {
		return fmt.Errorf("failed to authenticate with OpenStack: %w", err)
	}
	s.client = client
	return nil
}

func (s *openstackScraper) shutdown(_ context.Context) error {
	return nil
}

func (s *openstackScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	md := pmetric.NewMetrics()
	now := pcommon.NewTimestampFromTime(time.Now())

	var errs []error

	if err := s.scrapeNovaLimits(ctx, md, now); err != nil {
		errs = append(errs, err)
	}
	if s.cfg.QueryHypervisorMetrics {
		if err := s.scrapeHypervisors(ctx, md, now); err != nil {
			errs = append(errs, err)
		}
	}
	if err := s.scrapeNeutron(ctx, md, now); err != nil {
		errs = append(errs, err)
	}

	return md, errors.Join(errs...)
}

// commonAttrs returns the set of attributes shared by Nova-scoped metrics.
func (s *openstackScraper) commonAttrs() map[string]string {
	return map[string]string{
		"dsname":              "value",
		"plugin":              "openstack",
		"project_domain_name": s.client.ProjectDomainName(),
		"project_id":          s.client.ProjectID(),
		"project_name":        s.client.ProjectName(),
		"system.type":         "openstack",
		"user_domain_name":    s.client.UserDomainName(),
	}
}

// addGaugeInt appends a new ResourceMetrics entry with a single int gauge metric.
func addGaugeInt(md pmetric.Metrics, name string, value int64, ts pcommon.Timestamp, attrs map[string]string) {
	rm := md.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()
	m := sm.Metrics().AppendEmpty()
	m.SetName(name)
	dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
	dp.SetTimestamp(ts)
	dp.SetIntValue(value)
	for k, v := range attrs {
		dp.Attributes().PutStr(k, v)
	}
}

func (s *openstackScraper) scrapeNovaLimits(ctx context.Context, md pmetric.Metrics, now pcommon.Timestamp) error {
	limits, err := s.client.GetNovaLimits(ctx, s.client.ProjectID())
	if err != nil {
		return fmt.Errorf("failed to get Nova limits: %w", err)
	}

	abs := limits.Absolute
	attrs := s.commonAttrs()

	addGaugeInt(md, "gauge.openstack.nova.limit.maxImageMeta", int64(abs.MaxImageMeta), now, attrs)
	addGaugeInt(md, "gauge.openstack.nova.limit.maxSecurityGroups", int64(abs.MaxSecurityGroups), now, attrs)
	addGaugeInt(md, "gauge.openstack.nova.limit.maxTotalCores", int64(abs.MaxTotalCores), now, attrs)
	addGaugeInt(md, "gauge.openstack.nova.limit.maxTotalFloatingIps", int64(abs.MaxTotalFloatingIps), now, attrs)
	addGaugeInt(md, "gauge.openstack.nova.limit.maxTotalInstances", int64(abs.MaxTotalInstances), now, attrs)
	addGaugeInt(md, "gauge.openstack.nova.limit.maxTotalKeypairs", int64(abs.MaxTotalKeypairs), now, attrs)
	addGaugeInt(md, "gauge.openstack.nova.limit.maxTotalRAMSize", int64(abs.MaxTotalRAMSize), now, attrs)
	addGaugeInt(md, "gauge.openstack.nova.limit.totalCoresUsed", int64(abs.TotalCoresUsed), now, attrs)
	addGaugeInt(md, "gauge.openstack.nova.limit.totalFloatingIpsUsed", int64(abs.TotalFloatingIpsUsed), now, attrs)
	addGaugeInt(md, "gauge.openstack.nova.limit.totalInstancesUsed", int64(abs.TotalInstancesUsed), now, attrs)
	addGaugeInt(md, "gauge.openstack.nova.limit.totalRAMUsed", int64(abs.TotalRAMUsed), now, attrs)
	addGaugeInt(md, "gauge.openstack.nova.limit.totalSecurityGroupsUsed", int64(abs.TotalSecurityGroupsUsed), now, attrs)

	return nil
}

func (s *openstackScraper) scrapeHypervisors(ctx context.Context, md pmetric.Metrics, now pcommon.Timestamp) error {
	hvList, err := s.client.ListHypervisors(ctx)
	if err != nil {
		return fmt.Errorf("failed to list hypervisors: %w", err)
	}

	base := s.commonAttrs()

	for i := range hvList {
		hv := &hvList[i]

		// Build per-hypervisor attribute set.
		attrs := make(map[string]string, len(base)+7)
		maps.Copy(attrs, base)
		attrs["host_ip"] = hv.HostIP
		attrs["hypervisor_hostname"] = hv.HypervisorHostname
		attrs["hypervisor_type"] = hv.HypervisorType
		attrs["hypervisor_version"] = strconv.Itoa(hv.HypervisorVersion)
		attrs["id"] = hv.ID
		attrs["state"] = hv.State
		attrs["status"] = hv.Status

		addGaugeInt(md, "gauge.openstack.nova.hypervisor.current_workload", int64(hv.CurrentWorkload), now, attrs)
		addGaugeInt(md, "gauge.openstack.nova.hypervisor.disk_available_least", int64(hv.DiskAvailableLeast), now, attrs)
		addGaugeInt(md, "gauge.openstack.nova.hypervisor.free_disk_gb", int64(hv.FreeDiskGB), now, attrs)
		addGaugeInt(md, "gauge.openstack.nova.hypervisor.free_ram_mb", int64(hv.FreeRamMB), now, attrs)
		addGaugeInt(md, "gauge.openstack.nova.hypervisor.local_gb", int64(hv.LocalGB), now, attrs)
		addGaugeInt(md, "gauge.openstack.nova.hypervisor.local_gb_used", int64(hv.LocalGBUsed), now, attrs)
		addGaugeInt(md, "gauge.openstack.nova.hypervisor.memory_mb", int64(hv.MemoryMB), now, attrs)
		addGaugeInt(md, "gauge.openstack.nova.hypervisor.memory_mb_used", int64(hv.MemoryMBUsed), now, attrs)
		addGaugeInt(md, "gauge.openstack.nova.hypervisor.running_vms", int64(hv.RunningVMs), now, attrs)
		addGaugeInt(md, "gauge.openstack.nova.hypervisor.vcpus", int64(hv.VCPUs), now, attrs)
		addGaugeInt(md, "gauge.openstack.nova.hypervisor.vcpus_used", int64(hv.VCPUsUsed), now, attrs)
	}

	return nil
}

// neutronAttrs returns the set of attributes for Neutron metrics.
// Neutron reports global counts and does not include project_id.
func (s *openstackScraper) neutronAttrs() map[string]string {
	return map[string]string{
		"dsname":              "value",
		"plugin":              "openstack",
		"project_domain_name": s.client.ProjectDomainName(),
		"project_name":        s.client.ProjectName(),
		"system.type":         "openstack",
		"user_domain_name":    s.client.UserDomainName(),
	}
}

func (s *openstackScraper) scrapeNeutron(ctx context.Context, md pmetric.Metrics, now pcommon.Timestamp) error {
	attrs := s.neutronAttrs()
	var errs []error

	networkCount, err := s.client.CountNetworks(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to count networks: %w", err))
	} else {
		addGaugeInt(md, "gauge.openstack.neutron.network.count", int64(networkCount), now, attrs)
	}

	subnetCount, err := s.client.CountSubnets(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to count subnets: %w", err))
	} else {
		addGaugeInt(md, "gauge.openstack.neutron.subnet.count", int64(subnetCount), now, attrs)
	}

	routerCount, err := s.client.CountRouters(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to count routers: %w", err))
	} else {
		addGaugeInt(md, "gauge.openstack.neutron.router.count", int64(routerCount), now, attrs)
	}

	floatingIPCount, err := s.client.CountFloatingIPs(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to count floating IPs: %w", err))
	} else {
		addGaugeInt(md, "gauge.openstack.neutron.floatingip.count", int64(floatingIPCount), now, attrs)
	}

	sgCount, err := s.client.CountSecurityGroups(ctx)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to count security groups: %w", err))
	} else {
		addGaugeInt(md, "gauge.openstack.neutron.securitygroup.count", int64(sgCount), now, attrs)
	}

	return errors.Join(errs...)
}
