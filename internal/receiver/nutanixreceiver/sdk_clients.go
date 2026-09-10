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
	"io"
	"net/url"
	"time"

	clusterAPI "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/api"
	clusterClient "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/client"
	dataPoliciesAPI "github.com/nutanix/ntnx-api-golang-clients/datapolicies-go-client/v4/api"
	dataPoliciesClient "github.com/nutanix/ntnx-api-golang-clients/datapolicies-go-client/v4/client"
	dataProtectionAPI "github.com/nutanix/ntnx-api-golang-clients/dataprotection-go-client/v4/api"
	dataProtectionClient "github.com/nutanix/ntnx-api-golang-clients/dataprotection-go-client/v4/client"
	filesAPI "github.com/nutanix/ntnx-api-golang-clients/files-go-client/v4/api"
	filesClient "github.com/nutanix/ntnx-api-golang-clients/files-go-client/v4/client"
	microsegAPI "github.com/nutanix/ntnx-api-golang-clients/microseg-go-client/v4/api"
	microsegClient "github.com/nutanix/ntnx-api-golang-clients/microseg-go-client/v4/client"
	monitoringAPI "github.com/nutanix/ntnx-api-golang-clients/monitoring-go-client/v4/api"
	monitoringClient "github.com/nutanix/ntnx-api-golang-clients/monitoring-go-client/v4/client"
	networkingAPI "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/api"
	networkingClient "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/client"
	objectsAPI "github.com/nutanix/ntnx-api-golang-clients/objects-go-client/v4/api"
	objectsClient "github.com/nutanix/ntnx-api-golang-clients/objects-go-client/v4/client"
	prismAPI "github.com/nutanix/ntnx-api-golang-clients/prism-go-client/v4/api"
	prismSDKClient "github.com/nutanix/ntnx-api-golang-clients/prism-go-client/v4/client"
	vmmAPI "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/api"
	vmmClient "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/client"
	volumesAPI "github.com/nutanix/ntnx-api-golang-clients/volumes-go-client/v4/api"
	volumesClient "github.com/nutanix/ntnx-api-golang-clients/volumes-go-client/v4/client"
)

type prismClient struct {
	baseURL *url.URL

	clusters            *clusterAPI.ClustersServiceApi
	storageContainers   *clusterAPI.StorageContainersServiceApi
	disks               *clusterAPI.DisksServiceApi
	virtualMachines     *vmmAPI.VmServiceApi
	virtualMachineStats *vmmAPI.StatsServiceApi
	volumeGroups        *volumesAPI.VolumeGroupsServiceApi
	subnets             *networkingAPI.SubnetsServiceApi

	bgpSessions        *networkingAPI.BgpSessionsServiceApi
	gateways           *networkingAPI.GatewaysServiceApi
	layer2Stretches    *networkingAPI.Layer2StretchesServiceApi
	layer2StretchStats *networkingAPI.Layer2StretchStatsServiceApi
	networkControllers *networkingAPI.NetworkControllersServiceApi
	routingPolicies    *networkingAPI.RoutingPoliciesServiceApi
	trafficMirrors     *networkingAPI.TrafficMirrorsServiceApi
	trafficMirrorStats *networkingAPI.TrafficMirrorStatsServiceApi
	uplinkBonds        *networkingAPI.UplinkBondsServiceApi
	virtualSwitches    *networkingAPI.VirtualSwitchesServiceApi
	vpnConnections     *networkingAPI.VpnConnectionsServiceApi
	vpnConnectionStats *networkingAPI.VpnConnectionStatsServiceApi
	vpcs               *networkingAPI.VpcsServiceApi
	vpcStats           *networkingAPI.VpcNsStatsServiceApi

	categories         *prismAPI.CategoriesServiceApi
	tasks              *prismAPI.TasksServiceApi
	alerts             *monitoringAPI.AlertsServiceApi
	protectionPolicies *dataPoliciesAPI.ProtectionPoliciesServiceApi
	recoveryPoints     *dataProtectionAPI.RecoveryPointsServiceApi
	securityPolicies   *microsegAPI.NetworkSecurityPoliciesServiceApi
	addressGroups      *microsegAPI.AddressGroupsServiceApi
	serviceGroups      *microsegAPI.ServiceGroupsServiceApi

	fileServers       *filesAPI.FileServersApi
	unifiedNamespaces *filesAPI.UnifiedNamespacesApi
	antivirusServers  *filesAPI.AntivirusServersApi
	mountTargets      *filesAPI.MountTargetsApi
	filesAnalytics    *filesAPI.AnalyticsApi

	objectStores     *objectsAPI.ObjectStoresServiceApi
	objectStoreStats *objectsAPI.StatsServiceApi

	interval time.Duration
}

type generatedSDKClient interface {
	SetUserName(string)
	SetPassword(string)
	SetVerifySSL(bool)
	SetMaxRetryAttempts(int)
	SetRetryIntervalInMilliSeconds(int)
	SetLogOutput(io.Writer)
}

type generatedSDKFields struct {
	scheme         *string
	host           *string
	port           *int
	readTimeout    *time.Duration
	connectTimeout *time.Duration
}

func configureGeneratedSDKClient(client generatedSDKClient, fields generatedSDKFields, baseURL *url.URL, cfg *Config) {
	*fields.scheme = baseURL.Scheme
	*fields.host = baseURL.Hostname()
	*fields.port = int(serverPort(baseURL))
	*fields.readTimeout = cfg.ControllerConfig.Timeout
	*fields.connectTimeout = cfg.ControllerConfig.Timeout
	client.SetUserName(cfg.Username)
	client.SetPassword(string(cfg.Password))
	client.SetVerifySSL(!cfg.TLS.InsecureSkipVerify)
	client.SetMaxRetryAttempts(v4RequestAttempts - 1)
	client.SetRetryIntervalInMilliSeconds(250)
	client.SetLogOutput(io.Discard)
}

func newPrismV4Client(cfg *Config) (*prismClient, error) {
	baseURL, err := normalizeEndpoint(cfg.Endpoint, cfg.Port)
	if err != nil {
		return nil, err
	}

	clusterSDK := clusterClient.NewApiClient()
	configureGeneratedSDKClient(clusterSDK, generatedSDKFields{&clusterSDK.Scheme, &clusterSDK.Host, &clusterSDK.Port, &clusterSDK.ReadTimeout, &clusterSDK.ConnectTimeout}, baseURL, cfg)
	vmmSDK := vmmClient.NewApiClient()
	configureGeneratedSDKClient(vmmSDK, generatedSDKFields{&vmmSDK.Scheme, &vmmSDK.Host, &vmmSDK.Port, &vmmSDK.ReadTimeout, &vmmSDK.ConnectTimeout}, baseURL, cfg)
	volumesSDK := volumesClient.NewApiClient()
	configureGeneratedSDKClient(volumesSDK, generatedSDKFields{&volumesSDK.Scheme, &volumesSDK.Host, &volumesSDK.Port, &volumesSDK.ReadTimeout, &volumesSDK.ConnectTimeout}, baseURL, cfg)
	networkingSDK := networkingClient.NewApiClient()
	configureGeneratedSDKClient(networkingSDK, generatedSDKFields{&networkingSDK.Scheme, &networkingSDK.Host, &networkingSDK.Port, &networkingSDK.ReadTimeout, &networkingSDK.ConnectTimeout}, baseURL, cfg)
	prismSDK := prismSDKClient.NewApiClient()
	configureGeneratedSDKClient(prismSDK, generatedSDKFields{&prismSDK.Scheme, &prismSDK.Host, &prismSDK.Port, &prismSDK.ReadTimeout, &prismSDK.ConnectTimeout}, baseURL, cfg)
	monitoringSDK := monitoringClient.NewApiClient()
	configureGeneratedSDKClient(monitoringSDK, generatedSDKFields{&monitoringSDK.Scheme, &monitoringSDK.Host, &monitoringSDK.Port, &monitoringSDK.ReadTimeout, &monitoringSDK.ConnectTimeout}, baseURL, cfg)
	dataPoliciesSDK := dataPoliciesClient.NewApiClient()
	configureGeneratedSDKClient(dataPoliciesSDK, generatedSDKFields{&dataPoliciesSDK.Scheme, &dataPoliciesSDK.Host, &dataPoliciesSDK.Port, &dataPoliciesSDK.ReadTimeout, &dataPoliciesSDK.ConnectTimeout}, baseURL, cfg)
	dataProtectionSDK := dataProtectionClient.NewApiClient()
	configureGeneratedSDKClient(dataProtectionSDK, generatedSDKFields{&dataProtectionSDK.Scheme, &dataProtectionSDK.Host, &dataProtectionSDK.Port, &dataProtectionSDK.ReadTimeout, &dataProtectionSDK.ConnectTimeout}, baseURL, cfg)
	microsegSDK := microsegClient.NewApiClient()
	configureGeneratedSDKClient(microsegSDK, generatedSDKFields{&microsegSDK.Scheme, &microsegSDK.Host, &microsegSDK.Port, &microsegSDK.ReadTimeout, &microsegSDK.ConnectTimeout}, baseURL, cfg)
	filesSDK := filesClient.NewApiClient()
	configureGeneratedSDKClient(filesSDK, generatedSDKFields{&filesSDK.Scheme, &filesSDK.Host, &filesSDK.Port, &filesSDK.ReadTimeout, &filesSDK.ConnectTimeout}, baseURL, cfg)
	objectsSDK := objectsClient.NewApiClient()
	configureGeneratedSDKClient(objectsSDK, generatedSDKFields{&objectsSDK.Scheme, &objectsSDK.Host, &objectsSDK.Port, &objectsSDK.ReadTimeout, &objectsSDK.ConnectTimeout}, baseURL, cfg)

	return &prismClient{
		baseURL:             baseURL,
		interval:            cfg.ControllerConfig.CollectionInterval,
		clusters:            clusterAPI.NewClustersServiceApi(clusterSDK),
		storageContainers:   clusterAPI.NewStorageContainersServiceApi(clusterSDK),
		disks:               clusterAPI.NewDisksServiceApi(clusterSDK),
		virtualMachines:     vmmAPI.NewVmServiceApi(vmmSDK),
		virtualMachineStats: vmmAPI.NewStatsServiceApi(vmmSDK),
		volumeGroups:        volumesAPI.NewVolumeGroupsServiceApi(volumesSDK),
		subnets:             networkingAPI.NewSubnetsServiceApi(networkingSDK),
		bgpSessions:         networkingAPI.NewBgpSessionsServiceApi(networkingSDK),
		gateways:            networkingAPI.NewGatewaysServiceApi(networkingSDK),
		layer2Stretches:     networkingAPI.NewLayer2StretchesServiceApi(networkingSDK),
		layer2StretchStats:  networkingAPI.NewLayer2StretchStatsServiceApi(networkingSDK),
		networkControllers:  networkingAPI.NewNetworkControllersServiceApi(networkingSDK),
		routingPolicies:     networkingAPI.NewRoutingPoliciesServiceApi(networkingSDK),
		trafficMirrors:      networkingAPI.NewTrafficMirrorsServiceApi(networkingSDK),
		trafficMirrorStats:  networkingAPI.NewTrafficMirrorStatsServiceApi(networkingSDK),
		uplinkBonds:         networkingAPI.NewUplinkBondsServiceApi(networkingSDK),
		virtualSwitches:     networkingAPI.NewVirtualSwitchesServiceApi(networkingSDK),
		vpnConnections:      networkingAPI.NewVpnConnectionsServiceApi(networkingSDK),
		vpnConnectionStats:  networkingAPI.NewVpnConnectionStatsServiceApi(networkingSDK),
		vpcs:                networkingAPI.NewVpcsServiceApi(networkingSDK),
		vpcStats:            networkingAPI.NewVpcNsStatsServiceApi(networkingSDK),
		categories:          prismAPI.NewCategoriesServiceApi(prismSDK),
		tasks:               prismAPI.NewTasksServiceApi(prismSDK),
		alerts:              monitoringAPI.NewAlertsServiceApi(monitoringSDK),
		protectionPolicies:  dataPoliciesAPI.NewProtectionPoliciesServiceApi(dataPoliciesSDK),
		recoveryPoints:      dataProtectionAPI.NewRecoveryPointsServiceApi(dataProtectionSDK),
		securityPolicies:    microsegAPI.NewNetworkSecurityPoliciesServiceApi(microsegSDK),
		addressGroups:       microsegAPI.NewAddressGroupsServiceApi(microsegSDK),
		serviceGroups:       microsegAPI.NewServiceGroupsServiceApi(microsegSDK),
		fileServers:         filesAPI.NewFileServersApi(filesSDK),
		unifiedNamespaces:   filesAPI.NewUnifiedNamespacesApi(filesSDK),
		antivirusServers:    filesAPI.NewAntivirusServersApi(filesSDK),
		mountTargets:        filesAPI.NewMountTargetsApi(filesSDK),
		filesAnalytics:      filesAPI.NewAnalyticsApi(filesSDK),
		objectStores:        objectsAPI.NewObjectStoresServiceApi(objectsSDK),
		objectStoreStats:    objectsAPI.NewStatsServiceApi(objectsSDK),
	}, nil
}
