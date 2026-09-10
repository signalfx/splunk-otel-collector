// Copyright Splunk, Inc.
// Copyright The OpenTelemetry Authors
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

package nutanixreceiver

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	protectionPolicyRequests "github.com/nutanix/ntnx-api-golang-clients/datapolicies-go-client/v4/models/datapolicies/v4/request/protectionpolicies"
	recoveryPointRequests "github.com/nutanix/ntnx-api-golang-clients/dataprotection-go-client/v4/models/dataprotection/v4/request/recoverypoints"
	addressGroupRequests "github.com/nutanix/ntnx-api-golang-clients/microseg-go-client/v4/models/microseg/v4/request/addressgroups"
	securityPolicyRequests "github.com/nutanix/ntnx-api-golang-clients/microseg-go-client/v4/models/microseg/v4/request/networksecuritypolicies"
	serviceGroupRequests "github.com/nutanix/ntnx-api-golang-clients/microseg-go-client/v4/models/microseg/v4/request/servicegroups"
	alertRequests "github.com/nutanix/ntnx-api-golang-clients/monitoring-go-client/v4/models/monitoring/v4/request/alerts"
	networkStatsCommon "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/common/v1/stats"
	bgpSessionRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/bgpsessions"
	gatewayRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/gateways"
	layer2StretchRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/layer2stretches"
	layer2StretchStatsRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/layer2stretchstats"
	networkControllerRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/networkcontrollers"
	routingPolicyRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/routingpolicies"
	trafficMirrorRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/trafficmirrors"
	trafficMirrorStatsRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/trafficmirrorstats"
	uplinkBondRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/uplinkbonds"
	virtualSwitchRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/virtualswitches"
	vpcStatsRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/vpcnsstats"
	vpcRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/vpcs"
	vpnConnectionRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/vpnconnections"
	vpnConnectionStatsRequests "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/request/vpnconnectionstats"
	objectStatsCommon "github.com/nutanix/ntnx-api-golang-clients/objects-go-client/v4/models/common/v1/stats"
	objectStoreRequests "github.com/nutanix/ntnx-api-golang-clients/objects-go-client/v4/models/objects/v4/request/objectstores"
	objectStoreStatsRequests "github.com/nutanix/ntnx-api-golang-clients/objects-go-client/v4/models/objects/v4/request/stats"
	categoryRequests "github.com/nutanix/ntnx-api-golang-clients/prism-go-client/v4/models/prism/v4/request/categories"
	taskRequests "github.com/nutanix/ntnx-api-golang-clients/prism-go-client/v4/models/prism/v4/request/tasks"
)

type sdkStatsFetcher func(context.Context, map[string]any) (any, error)

type sdkEntityEndpoint struct {
	list       sdkPageFetcher
	stats      sdkStatsFetcher
	entityType string
}

type sdkFilteredPageFetcher func(context.Context, int, int, string) (any, error)

func (c *prismClient) collectAdditionalMetrics(ctx context.Context, request additionalMetricsRequest) (additionalSnapshot, error) {
	var result additionalSnapshot
	collectors := []struct {
		collect func(context.Context) additionalSnapshot
		enabled bool
	}{
		{collect: c.collectNetworking, enabled: request.Networking},
		{collect: c.collectPrismCentral, enabled: request.PrismCentral},
		{collect: func(ctx context.Context) additionalSnapshot {
			return c.collectDataProtection(ctx, request.VMs)
		}, enabled: request.DataProtection},
		{collect: c.collectMicrosegmentation, enabled: request.Microsegmentation},
		{collect: c.collectFiles, enabled: request.Files},
		{collect: c.collectObjects, enabled: request.Objects},
	}
	for _, collector := range collectors {
		if !collector.enabled {
			continue
		}
		if err := ctx.Err(); err != nil {
			return result, err
		}
		partial := collector.collect(ctx)
		result.Metrics = append(result.Metrics, partial.Metrics...)
		result.Errors = append(result.Errors, partial.Errors...)
		if err := ctx.Err(); err != nil {
			return result, err
		}
	}
	return result, nil
}

func (c *prismClient) collectNetworking(ctx context.Context) additionalSnapshot {
	endpoints := []sdkEntityEndpoint{
		{entityType: "bgp_session", list: func(ctx context.Context, page, limit int) (any, error) {
			return c.bgpSessions.ListBgpSessions(ctx, &bgpSessionRequests.ListBgpSessionsRequest{Page_: &page, Limit_: &limit})
		}},
		{entityType: "gateway", list: func(ctx context.Context, page, limit int) (any, error) {
			return c.gateways.ListGateways(ctx, &gatewayRequests.ListGatewaysRequest{Page_: &page, Limit_: &limit})
		}},
		{entityType: "layer2_stretch", list: func(ctx context.Context, page, limit int) (any, error) {
			return c.layer2Stretches.ListLayer2Stretches(ctx, &layer2StretchRequests.ListLayer2StretchesRequest{Page_: &page, Limit_: &limit})
		}, stats: c.getLayer2StretchStats},
		{entityType: "network_controller", list: func(ctx context.Context, page, limit int) (any, error) {
			return c.networkControllers.ListNetworkControllers(ctx, &networkControllerRequests.ListNetworkControllersRequest{Page_: &page, Limit_: &limit})
		}},
		{entityType: "routing_policy", list: func(ctx context.Context, page, limit int) (any, error) {
			return c.routingPolicies.ListRoutingPolicies(ctx, &routingPolicyRequests.ListRoutingPoliciesRequest{Page_: &page, Limit_: &limit})
		}},
		{entityType: "traffic_mirror", list: func(ctx context.Context, page, limit int) (any, error) {
			return c.trafficMirrors.ListTrafficMirrors(ctx, &trafficMirrorRequests.ListTrafficMirrorsRequest{Page_: &page, Limit_: &limit})
		}, stats: c.getTrafficMirrorStats},
		{entityType: "uplink_bond", list: func(ctx context.Context, page, limit int) (any, error) {
			return c.uplinkBonds.ListUplinkBonds(ctx, &uplinkBondRequests.ListUplinkBondsRequest{Page_: &page, Limit_: &limit})
		}},
		{entityType: "virtual_switch", list: func(ctx context.Context, page, limit int) (any, error) {
			return c.virtualSwitches.ListVirtualSwitches(ctx, &virtualSwitchRequests.ListVirtualSwitchesRequest{Page_: &page, Limit_: &limit})
		}},
		{entityType: "vpn_connection", list: func(ctx context.Context, page, limit int) (any, error) {
			return c.vpnConnections.ListVpnConnections(ctx, &vpnConnectionRequests.ListVpnConnectionsRequest{Page_: &page, Limit_: &limit})
		}, stats: c.getVPNConnectionStats},
	}

	var result additionalSnapshot
	for _, endpoint := range endpoints {
		entities, err := listSDKEntities(ctx, "networking "+endpoint.entityType, endpoint.list)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("list networking %s: %w", endpoint.entityType, err))
			continue
		}
		result.Metrics = append(result.Metrics, inventoryCountMetric("networking", endpoint.entityType, "", "", len(entities)))
		if endpoint.stats == nil {
			continue
		}
		stats, errs := c.collectEntityStats(ctx, "networking", endpoint.entityType, entities, endpoint.stats)
		result.Metrics = append(result.Metrics, stats...)
		result.Errors = append(result.Errors, errs...)
	}

	vpcs, err := listSDKEntities(ctx, "networking vpcs", func(ctx context.Context, page, limit int) (any, error) {
		return c.vpcs.ListVpcs(ctx, &vpcRequests.ListVpcsRequest{Page_: &page, Limit_: &limit})
	})
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list networking vpc: %w", err))
		return result
	}
	result.Metrics = append(result.Metrics, inventoryCountMetric("networking", "vpc", "", "", len(vpcs)))
	var externalSubnets []map[string]any
	for _, vpc := range vpcs {
		vpcID := firstString(vpc, "extId", "ext_id", "id")
		vpcName := firstString(vpc, "name")
		for _, subnet := range mapList(vpc, "externalSubnets", "external_subnets") {
			subnet["vpcExtId"] = vpcID
			subnet["vpcName"] = vpcName
			externalSubnets = append(externalSubnets, subnet)
		}
	}
	stats, errs := c.collectEntityStats(ctx, "networking", "vpc_external_subnet", externalSubnets, c.getVPCExternalSubnetStats)
	result.Metrics = append(result.Metrics, stats...)
	result.Errors = append(result.Errors, errs...)
	return result
}

func (c *prismClient) getLayer2StretchStats(ctx context.Context, entity map[string]any) (any, error) {
	start, end, sampling := c.statQuery()
	statType := networkStatsCommon.DOWNSAMPLINGOPERATOR_LAST
	id := firstString(entity, "extId", "ext_id", "id")
	page, limit := 0, 1
	return c.layer2StretchStats.GetLayer2StretchStats(ctx, &layer2StretchStatsRequests.GetLayer2StretchStatsRequest{
		ExtId: &id, StartTime_: &start, EndTime_: &end, SamplingInterval_: &sampling, StatType_: &statType, Page_: &page, Limit_: &limit,
	})
}

func (c *prismClient) getTrafficMirrorStats(ctx context.Context, entity map[string]any) (any, error) {
	start, end, sampling := c.statQuery()
	statType := networkStatsCommon.DOWNSAMPLINGOPERATOR_LAST
	id := firstString(entity, "extId", "ext_id", "id")
	return c.trafficMirrorStats.GetTrafficMirrorStats(ctx, &trafficMirrorStatsRequests.GetTrafficMirrorStatsRequest{
		ExtId: &id, StartTime_: &start, EndTime_: &end, SamplingInterval_: &sampling, StatType_: &statType,
	})
}

func (c *prismClient) getVPNConnectionStats(ctx context.Context, entity map[string]any) (any, error) {
	start, end, sampling := c.statQuery()
	statType := networkStatsCommon.DOWNSAMPLINGOPERATOR_LAST
	id := firstString(entity, "extId", "ext_id", "id")
	page, limit := 0, 1
	return c.vpnConnectionStats.GetVpnConnectionStats(ctx, &vpnConnectionStatsRequests.GetVpnConnectionStatsRequest{
		ExtId: &id, StartTime_: &start, EndTime_: &end, SamplingInterval_: &sampling, StatType_: &statType, Page_: &page, Limit_: &limit,
	})
}

func (c *prismClient) getVPCExternalSubnetStats(ctx context.Context, entity map[string]any) (any, error) {
	start, end, sampling := c.statQuery()
	statType := networkStatsCommon.DOWNSAMPLINGOPERATOR_LAST
	vpcID := firstString(entity, "vpcExtId")
	subnetID := firstString(entity, "subnetReference", "subnet_reference", "extId")
	page, limit := 0, 1
	return c.vpcStats.GetVpcNsStats(ctx, &vpcStatsRequests.GetVpcNsStatsRequest{
		VpcExtId: &vpcID, ExtId: &subnetID, StartTime_: &start, EndTime_: &end, SamplingInterval_: &sampling, StatType_: &statType, Page_: &page, Limit_: &limit,
	})
}

func (c *prismClient) collectPrismCentral(ctx context.Context) additionalSnapshot {
	var result additionalSnapshot

	categories, err := listSDKEntities(ctx, "categories", func(ctx context.Context, page, limit int) (any, error) {
		return c.categories.ListCategories(ctx, &categoryRequests.ListCategoriesRequest{Page_: &page, Limit_: &limit})
	})
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list categories: %w", err))
	} else {
		result.Metrics = append(result.Metrics, inventoryCountMetric("prism", "category", "", "", len(categories)))
		for _, categoryType := range []string{"system", "user", "internal"} {
			result.Metrics = append(result.Metrics, inventoryCountMetric("prism", "category", "type", categoryType, countMapsByString(categories, "type", categoryType)))
		}
		keys := map[string]struct{}{}
		for _, category := range categories {
			if key := firstString(category, "key"); key != "" {
				keys[key] = struct{}{}
			}
		}
		result.Metrics = append(result.Metrics, inventoryCountMetric("prism", "category_key", "", "", len(keys)))
	}

	tasks := func(ctx context.Context, page, limit int, filter string) (any, error) {
		request := &taskRequests.ListTasksRequest{Page_: &page, Limit_: &limit}
		if filter != "" {
			request.Filter_ = &filter
		}
		return c.tasks.ListTasks(ctx, request)
	}
	result.appendSDKCount(ctx, "prism", "task", "", "", "", tasks)
	for _, status := range []string{"queued", "running", "canceling", "succeeded", "failed", "canceled", "suspended"} {
		filter := fmt.Sprintf("status eq Prism.Config.TaskStatus'%s'", strings.ToUpper(status))
		result.appendSDKCount(ctx, "prism", "task", "status", status, filter, tasks)
	}

	alerts := func(ctx context.Context, page, limit int, filter string) (any, error) {
		request := &alertRequests.ListAlertsRequest{Page_: &page, Limit_: &limit}
		if filter != "" {
			request.Filter_ = &filter
		}
		return c.alerts.ListAlerts(ctx, request)
	}
	result.appendSDKCount(ctx, "monitoring", "alert", "", "", "", alerts)
	for _, resolved := range []bool{true, false} {
		state := strconv.FormatBool(resolved)
		result.appendSDKCount(ctx, "monitoring", "alert", "resolved", state, "isResolved eq "+state, alerts)
	}
	for _, acknowledged := range []bool{true, false} {
		state := strconv.FormatBool(acknowledged)
		result.appendSDKCount(ctx, "monitoring", "alert", "acknowledged", state, "isAcknowledged eq "+state, alerts)
	}
	for _, severity := range []string{"info", "warning", "critical"} {
		severityFilter := fmt.Sprintf("severity eq Monitoring.Common.Severity'%s'", strings.ToUpper(severity))
		result.appendSDKCount(ctx, "monitoring", "alert", "severity", severity, severityFilter, alerts)
		result.appendSDKCount(ctx, "monitoring", "alert", "unresolved_severity", severity, "isResolved eq false and "+severityFilter, alerts)
	}
	return result
}

func (c *prismClient) collectDataProtection(ctx context.Context, vms []nutanixVM) additionalSnapshot {
	var result additionalSnapshot
	policies, err := listSDKEntities(ctx, "protection policies", func(ctx context.Context, page, limit int) (any, error) {
		return c.protectionPolicies.ListProtectionPolicies(ctx, &protectionPolicyRequests.ListProtectionPoliciesRequest{Page_: &page, Limit_: &limit})
	})
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list protection policies: %w", err))
	} else {
		result.Metrics = append(result.Metrics, inventoryCountMetric("data_protection", "protection_policy", "", "", len(policies)))
		counts := countProtectionPolicySchedules(policies)
		result.Metrics = append(result.Metrics, inventoryCountMetric("data_protection", "protection_policy_schedule", "", "", counts["total"]))
		for _, consistency := range []string{"crash_consistent", "application_consistent"} {
			result.Metrics = append(result.Metrics, inventoryCountMetric("data_protection", "protection_policy_schedule", "consistency", consistency, counts[consistency]))
		}
		for _, rpo := range []string{"sync", "nearsync", "async"} {
			result.Metrics = append(result.Metrics, inventoryCountMetric("data_protection", "protection_policy_schedule", "rpo", rpo, counts[rpo]))
		}
		protectedVMsByRPO := countProtectedVMsByRPO(policies, vms)
		for _, rpo := range []string{"sync", "nearsync", "async"} {
			result.Metrics = append(result.Metrics, inventoryCountMetric("data_protection", "protected_vm", "rpo", rpo, protectedVMsByRPO[rpo]))
		}
	}

	recoveryPoints, err := countSDKEntities(ctx, "recovery points", "", func(ctx context.Context, page, limit int, _ string) (any, error) {
		return c.recoveryPoints.ListRecoveryPoints(ctx, &recoveryPointRequests.ListRecoveryPointsRequest{Page_: &page, Limit_: &limit})
	})
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list recovery points: %w", err))
	} else {
		result.Metrics = append(result.Metrics, inventoryCountMetric("data_protection", "recovery_point", "", "", recoveryPoints))
	}
	return result
}

func (c *prismClient) collectMicrosegmentation(ctx context.Context) additionalSnapshot {
	var result additionalSnapshot
	policies, err := listSDKEntities(ctx, "microsegmentation policies", func(ctx context.Context, page, limit int) (any, error) {
		return c.securityPolicies.ListNetworkSecurityPolicies(ctx, &securityPolicyRequests.ListNetworkSecurityPoliciesRequest{Page_: &page, Limit_: &limit})
	})
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list microsegmentation policies: %w", err))
	} else {
		result.Metrics = append(result.Metrics,
			inventoryCountMetric("microseg", "network_security_policy", "", "", len(policies)),
			inventoryCountMetric("microseg", "network_security_policy", "scope", "vlan", countMicrosegPoliciesByScope(policies, "vlan", "all_vlan")),
			inventoryCountMetric("microseg", "network_security_policy", "scope", "vpc", countMicrosegPoliciesByScope(policies, "vpc", "all_vpc", "vpc_list")),
		)
		for _, state := range []string{"save", "monitor", "enforce"} {
			result.Metrics = append(result.Metrics, inventoryCountMetric("microseg", "network_security_policy", "state", state, countMapsByString(policies, "state", state)))
		}
		for _, policyType := range []string{"quarantine", "isolation", "application"} {
			result.Metrics = append(result.Metrics, inventoryCountMetric("microseg", "network_security_policy", "type", policyType, countMapsByString(policies, "type", policyType)))
		}
	}

	counts := []struct {
		fetch      sdkFilteredPageFetcher
		entityType string
	}{
		{entityType: "address_group", fetch: func(ctx context.Context, page, limit int, _ string) (any, error) {
			return c.addressGroups.ListAddressGroups(ctx, &addressGroupRequests.ListAddressGroupsRequest{Page_: &page, Limit_: &limit})
		}},
		{entityType: "service_group", fetch: func(ctx context.Context, page, limit int, _ string) (any, error) {
			return c.serviceGroups.ListServiceGroups(ctx, &serviceGroupRequests.ListServiceGroupsRequest{Page_: &page, Limit_: &limit})
		}},
	}
	for _, endpoint := range counts {
		count, countErr := countSDKEntities(ctx, "microsegmentation "+endpoint.entityType, "", endpoint.fetch)
		if countErr != nil {
			result.Errors = append(result.Errors, fmt.Errorf("list microsegmentation %s: %w", endpoint.entityType, countErr))
			continue
		}
		result.Metrics = append(result.Metrics, inventoryCountMetric("microseg", endpoint.entityType, "", "", count))
	}
	return result
}

func (c *prismClient) collectFiles(ctx context.Context) additionalSnapshot {
	var result additionalSnapshot
	fileServers, err := listSDKEntities(ctx, "file servers", func(ctx context.Context, page, limit int) (any, error) {
		return legacySDKCall(ctx, func() (any, error) {
			return c.fileServers.ListFileServers(&page, &limit, nil, nil, nil)
		})
	})
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list file servers: %w", err))
	} else {
		result.Metrics = append(result.Metrics, inventoryCountMetric("files", "file_server", "", "", len(fileServers)))
		stats, errs := c.collectEntityStats(ctx, "files", "file_server", fileServers, c.getFileServerStats)
		result.Metrics = append(result.Metrics, stats...)
		result.Errors = append(result.Errors, errs...)
	}

	unifiedNamespaces, countErr := countSDKEntities(ctx, "unified namespaces", "", func(ctx context.Context, page, limit int, _ string) (any, error) {
		return legacySDKCall(ctx, func() (any, error) {
			return c.unifiedNamespaces.ListUnifiedNamespaces(&page, &limit, nil, nil)
		})
	})
	if countErr != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list unified namespaces: %w", countErr))
	} else {
		result.Metrics = append(result.Metrics, inventoryCountMetric("files", "unified_namespace", "", "", unifiedNamespaces))
	}

	for _, fileServer := range fileServers {
		fileServerID := firstString(fileServer, "extId", "ext_id", "id")
		if fileServerID == "" {
			continue
		}
		nested := []sdkEntityEndpoint{
			{entityType: "antivirus_server", list: func(ctx context.Context, page, limit int) (any, error) {
				return legacySDKCall(ctx, func() (any, error) {
					return c.antivirusServers.ListAntivirusServers(&fileServerID, &page, &limit, nil, nil, nil)
				})
			}, stats: func(ctx context.Context, entity map[string]any) (any, error) {
				return c.getAntivirusServerStats(ctx, fileServerID, entity)
			}},
			{entityType: "mount_target", list: func(ctx context.Context, page, limit int) (any, error) {
				return legacySDKCall(ctx, func() (any, error) {
					return c.mountTargets.ListMountTargets(&fileServerID, &page, &limit, nil, nil, nil)
				})
			}, stats: func(ctx context.Context, entity map[string]any) (any, error) {
				return c.getMountTargetStats(ctx, fileServerID, entity)
			}},
		}
		for _, endpoint := range nested {
			entities, listErr := listSDKEntities(ctx, "files "+endpoint.entityType, endpoint.list)
			if listErr != nil {
				result.Errors = append(result.Errors, fmt.Errorf("list files %s for %s: %w", endpoint.entityType, fileServerID, listErr))
				continue
			}
			countMetric := inventoryCountMetric("files", endpoint.entityType, "file_server", fileServerID, len(entities))
			countMetric.Attributes["nutanix.files.file_server.name"] = firstString(fileServer, "name")
			result.Metrics = append(result.Metrics, countMetric)
			for _, entity := range entities {
				entity["fileServerID"] = fileServerID
				entity["fileServerName"] = firstString(fileServer, "name")
			}
			stats, errs := c.collectEntityStats(ctx, "files", endpoint.entityType, entities, endpoint.stats)
			result.Metrics = append(result.Metrics, stats...)
			result.Errors = append(result.Errors, errs...)
		}
	}
	return result
}

func (c *prismClient) fileStatsRange() (time.Time, time.Time, int) {
	interval := c.interval
	if interval < 5*time.Minute {
		interval = 5 * time.Minute
	}
	end := time.Now()
	return end.Add(-interval), end, 300
}

func (c *prismClient) getFileServerStats(ctx context.Context, entity map[string]any) (any, error) {
	start, end, sampling := c.fileStatsRange()
	id := firstString(entity, "extId", "ext_id", "id")
	return legacySDKCall(ctx, func() (any, error) {
		return c.filesAnalytics.GetFileServerStats(&id, &start, &end, &sampling, nil, nil)
	})
}

func (c *prismClient) getAntivirusServerStats(ctx context.Context, fileServerID string, entity map[string]any) (any, error) {
	start, end, sampling := c.fileStatsRange()
	id := firstString(entity, "extId", "ext_id", "id")
	return legacySDKCall(ctx, func() (any, error) {
		return c.filesAnalytics.GetAntivirusServerStats(&fileServerID, &id, &start, &end, &sampling, nil, nil)
	})
}

func (c *prismClient) getMountTargetStats(ctx context.Context, fileServerID string, entity map[string]any) (any, error) {
	start, end, sampling := c.fileStatsRange()
	id := firstString(entity, "extId", "ext_id", "id")
	return legacySDKCall(ctx, func() (any, error) {
		return c.filesAnalytics.GetMountTargetStats(&fileServerID, &id, &start, &end, &sampling, nil, nil)
	})
}

func (c *prismClient) collectObjects(ctx context.Context) additionalSnapshot {
	var result additionalSnapshot
	objectStores, err := listSDKEntities(ctx, "object stores", func(ctx context.Context, page, limit int) (any, error) {
		return c.objectStores.ListObjectstores(ctx, &objectStoreRequests.ListObjectstoresRequest{Page_: &page, Limit_: &limit})
	})
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("list object stores: %w", err))
		return result
	}
	result.Metrics = append(result.Metrics, inventoryCountMetric("objects", "object_store", "", "", len(objectStores)))
	stats, errs := c.collectEntityStats(ctx, "objects", "object_store", objectStores, c.getObjectStoreStats)
	result.Metrics = append(result.Metrics, stats...)
	result.Errors = append(result.Errors, errs...)
	return result
}

func (c *prismClient) getObjectStoreStats(ctx context.Context, entity map[string]any) (any, error) {
	interval := c.interval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	sampling := int(math.Max(120, interval.Seconds()))
	statType := objectStatsCommon.DOWNSAMPLINGOPERATOR_LAST
	end := time.Now()
	start := end.Add(-interval)
	id := firstString(entity, "extId", "ext_id", "id")
	return c.objectStoreStats.GetObjectstoreStatsById(ctx, &objectStoreStatsRequests.GetObjectstoreStatsByIdRequest{
		ExtId: &id, StartTime_: &start, EndTime_: &end, SamplingInterval_: &sampling, StatType_: &statType,
	})
}

func (c *prismClient) collectEntityStats(
	ctx context.Context,
	domain string,
	entityType string,
	entities []map[string]any,
	fetch sdkStatsFetcher,
) ([]additionalMetric, []error) {
	metrics := make([]additionalMetric, len(entities))
	errorsByIndex := make([]error, len(entities))
	sem := make(chan struct{}, v4StatsWorkers)
	var wg sync.WaitGroup
	for i := range entities {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				errorsByIndex[index] = ctx.Err()
				return
			}
			entity := entities[index]
			entityID := firstString(entity, "extId", "ext_id", "id", "subnetReference", "subnet_reference")
			if entityID == "" {
				return
			}
			response, err := fetch(ctx, entity)
			stats, err := statsFromSDKResponse(response, err)
			if err != nil {
				errorsByIndex[index] = fmt.Errorf("get %s %s stats for %s: %w", domain, entityType, entityID, err)
				return
			}
			metricPrefix := inventoryMetricPrefix(domain, entityType)
			attributes := map[string]string{
				metricPrefix + ".id":   entityID,
				metricPrefix + ".name": firstString(entity, "name"),
			}
			if domain == "networking" && entityType == "vpc_external_subnet" {
				attributes["nutanix.networking.vpc.id"] = firstString(entity, "vpcExtId")
				attributes["nutanix.networking.vpc.name"] = firstString(entity, "vpcName")
			}
			if domain == "files" && entityType != "file_server" {
				attributes["nutanix.files.file_server.id"] = firstString(entity, "fileServerID")
				attributes["nutanix.files.file_server.name"] = firstString(entity, "fileServerName")
			}
			metrics[index] = additionalMetric{
				Name:        metricPrefix + ".stat",
				Description: entityStatsDescription(domain, entityType),
				Attributes:  attributes,
				Stats:       stats,
			}
		}(i)
	}
	wg.Wait()
	return compactAdditionalMetrics(metrics), compactErrors(errorsByIndex)
}

func entityStatsDescription(domain, entityType string) string {
	switch {
	case domain == "files" && entityType == "antivirus_server":
		return "Latest statistic for a Nutanix Files antivirus server."
	case domain == "files" && entityType == "file_server":
		return "Latest statistic for a Nutanix Files file server."
	case domain == "files" && entityType == "mount_target":
		return "Latest statistic for a Nutanix Files mount target."
	case domain == "networking" && entityType == "layer2_stretch":
		return "Latest statistic for a Nutanix Layer 2 stretch."
	case domain == "networking" && entityType == "traffic_mirror":
		return "Latest statistic for a Nutanix traffic mirror."
	case domain == "networking" && entityType == "vpc_external_subnet":
		return "Latest statistic for an external subnet attached to a Nutanix VPC."
	case domain == "networking" && entityType == "vpn_connection":
		return "Latest statistic for a Nutanix VPN connection."
	case domain == "objects" && entityType == "object_store":
		return "Latest statistic for a Nutanix Objects object store."
	default:
		return "Latest statistic for a Nutanix " + strings.ReplaceAll(entityType, "_", " ") + "."
	}
}

func countSDKEntities(ctx context.Context, name, filter string, fetch sdkFilteredPageFetcher) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	response, err := fetch(ctx, 0, 1, filter)
	if err != nil {
		return 0, err
	}
	responseMap, err := sdkResponseMap(response)
	if err != nil {
		return 0, fmt.Errorf("decode %s response: %w", name, err)
	}
	if total, ok := responseTotal(responseMap); ok {
		return total, nil
	}
	return len(responseDataMaps(responseMap)), nil
}

func (s *additionalSnapshot) appendSDKCount(ctx context.Context, domain, entityType, stateType, state, filter string, fetch sdkFilteredPageFetcher) {
	count, err := countSDKEntities(ctx, domain+" "+entityType, filter, fetch)
	if err != nil {
		s.Errors = append(s.Errors, fmt.Errorf("count %s %s: %w", domain, entityType, err))
		return
	}
	s.Metrics = append(s.Metrics, inventoryCountMetric(domain, entityType, stateType, state, count))
}

func legacySDKCall[T any](ctx context.Context, call func() (T, error)) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	result, err := call()
	if err != nil {
		return zero, err
	}
	if err := ctx.Err(); err != nil {
		return zero, err
	}
	return result, nil
}
