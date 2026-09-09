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

// Code generated from components.go and go.mod by "go generate"; DO NOT EDIT.
// Production versions are resolved from the final binary's Go build info so
// metadata follows Minimal Version Selection in consuming flavors. The pinned
// references are fallbacks only for dependency-less non-main test binaries.

package baseline

var baselineExtensionModules = []moduleMetadata{
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/ackextension", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/ackextension v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/basicauthextension", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/basicauthextension v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/bearertokenauthextension", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/bearertokenauthextension v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/dbauth/awsiamdbauthextension", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/dbauth/awsiamdbauthextension v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/dockerobserver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/dockerobserver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecsobserver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecsobserver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/googlecloudlogentryencodingextension", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/googlecloudlogentryencodingextension v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/headerssetterextension", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/headerssetterextension v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarderextension", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarderextension v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/k8sleaderelector", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/k8sleaderelector v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/oauth2clientauthextension", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/oauth2clientauthextension v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampextension", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampextension v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/textencodingextension", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/textencodingextension v0.161.0"},
	{path: "go.opentelemetry.io/collector/extension/zpagesextension", fallback: "go.opentelemetry.io/collector/extension/zpagesextension v0.161.0"},
}

var baselineReceiverModules = []moduleMetadata{
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/activedirectorydsreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/activedirectorydsreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachesparkreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachesparkreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscontainerinsightreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscontainerinsightreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureblobreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureblobreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureeventhubreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureeventhubreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azuremonitorreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azuremonitorreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/chronyreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/chronyreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/ciscoosreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/ciscoosreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudfoundryreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudfoundryreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dnscheckreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dnscheckreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filestatsreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filestatsreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudpubsubreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudpubsubreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/haproxyreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/haproxyreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/httpcheckreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/httpcheckreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/icmpcheckreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/icmpcheckreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/iisreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/iisreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/influxdbreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/influxdbreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/memcachedreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/memcachedreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver v0.161.0"},
	{path: "go.opentelemetry.io/collector/receiver/nopreceiver", fallback: "go.opentelemetry.io/collector/receiver/nopreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/ntpreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/ntpreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/oracledbreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/oracledbreceiver v0.161.0"},
	{path: "go.opentelemetry.io/collector/receiver/otlpreceiver", fallback: "go.opentelemetry.io/collector/receiver/otlpreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusremotewritereceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusremotewritereceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/purefareceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/purefareceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/rabbitmqreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/rabbitmqreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/saphanareceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/saphanareceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/snmpreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/snmpreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/snowflakereceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/snowflakereceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/solacereceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/solacereceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkenterprisereceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkenterprisereceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlqueryreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlqueryreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlserverreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlserverreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sshcheckreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sshcheckreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/systemdreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/systemdreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcpcheckreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcpcheckreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcplogreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcplogreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tlscheckreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tlscheckreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/udplogreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/udplogreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/wavefrontreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/wavefrontreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsservicereceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsservicereceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/yanggrpcreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/yanggrpcreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zookeeperreceiver", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zookeeperreceiver v0.161.0"},
}

var baselineProcessorModules = []moduleMetadata{
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.161.0"},
	{path: "go.opentelemetry.io/collector/processor/batchprocessor", fallback: "go.opentelemetry.io/collector/processor/batchprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor v1.0.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/lookupprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/lookupprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/logstransformprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/logstransformprocessor v0.161.0"},
	{path: "go.opentelemetry.io/collector/processor/memorylimiterprocessor", fallback: "go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricsgenerationprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricsgenerationprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/redactionprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/redactionprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.161.0"},
}

var baselineExporterModules = []moduleMetadata{
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awss3exporter", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awss3exporter v0.161.0"},
	{path: "go.opentelemetry.io/collector/exporter/debugexporter", fallback: "go.opentelemetry.io/collector/exporter/debugexporter v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudstorageexporter", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudstorageexporter v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter v0.161.0"},
	{path: "go.opentelemetry.io/collector/exporter/nopexporter", fallback: "go.opentelemetry.io/collector/exporter/nopexporter v0.161.0"},
	{path: "go.opentelemetry.io/collector/exporter/otlpexporter", fallback: "go.opentelemetry.io/collector/exporter/otlpexporter v0.161.0"},
	{path: "go.opentelemetry.io/collector/exporter/otlphttpexporter", fallback: "go.opentelemetry.io/collector/exporter/otlphttpexporter v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter v0.161.0"},
}

var baselineConnectorModules = []moduleMetadata{
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/connector/countconnector", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/connector/countconnector v0.161.0"},
	{path: "go.opentelemetry.io/collector/connector/forwardconnector", fallback: "go.opentelemetry.io/collector/connector/forwardconnector v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/connector/routingconnector", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/connector/routingconnector v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector v0.161.0"},
	{path: "github.com/open-telemetry/opentelemetry-collector-contrib/connector/sumconnector", fallback: "github.com/open-telemetry/opentelemetry-collector-contrib/connector/sumconnector v0.161.0"},
}
