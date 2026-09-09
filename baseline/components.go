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

// The baseline component set. Hand-maintained until a follow-up adds the
// generator that produces this file from an ocb manifest.

package baseline

import (
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"

	countconnector "github.com/open-telemetry/opentelemetry-collector-contrib/connector/countconnector"
	routingconnector "github.com/open-telemetry/opentelemetry-collector-contrib/connector/routingconnector"
	spanmetricsconnector "github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector"
	sumconnector "github.com/open-telemetry/opentelemetry-collector-contrib/connector/sumconnector"
	awss3exporter "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awss3exporter"
	fileexporter "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter"
	googlecloudstorageexporter "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudstorageexporter"
	kafkaexporter "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter"
	loadbalancingexporter "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter"
	prometheusremotewriteexporter "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter"
	signalfxexporter "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/signalfxexporter"
	splunkhecexporter "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter"
	ackextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/ackextension"
	basicauthextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/basicauthextension"
	bearertokenauthextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/bearertokenauthextension"
	awsiamdbauthextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/dbauth/awsiamdbauthextension"
	googlecloudlogentryencodingextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/googlecloudlogentryencodingextension"
	textencodingextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/textencodingextension"
	headerssetterextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/headerssetterextension"
	healthcheckextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	httpforwarderextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/httpforwarderextension"
	k8sleaderelector "github.com/open-telemetry/opentelemetry-collector-contrib/extension/k8sleaderelector"
	oauth2clientauthextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/oauth2clientauthextension"
	dockerobserver "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/dockerobserver"
	ecsobserver "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecsobserver"
	hostobserver "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/hostobserver"
	k8sobserver "github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/k8sobserver"
	opampextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampextension"
	pprofextension "github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension"
	filestorage "github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage"
	attributesprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	cumulativetodeltaprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor"
	filterprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	groupbyattrsprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor"
	k8sattributesprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"
	logstransformprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/logstransformprocessor"
	metricsgenerationprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricsgenerationprocessor"
	metricstransformprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"
	probabilisticsamplerprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor"
	redactionprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/redactionprocessor"
	resourcedetectionprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	resourceprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	spanprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/spanprocessor"
	tailsamplingprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor"
	transformprocessor "github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	activedirectorydsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/activedirectorydsreceiver"
	apachereceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver"
	apachesparkreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachesparkreceiver"
	awscloudwatchreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchreceiver"
	awscontainerinsightreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscontainerinsightreceiver"
	awsecscontainermetricsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver"
	azureblobreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureblobreceiver"
	azureeventhubreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureeventhubreceiver"
	azuremonitorreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azuremonitorreceiver"
	carbonreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver"
	chronyreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/chronyreceiver"
	ciscoosreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/ciscoosreceiver"
	cloudfoundryreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudfoundryreceiver"
	collectdreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver"
	dnscheckreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dnscheckreceiver"
	dockerstatsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver"
	elasticsearchreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver"
	filelogreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	filestatsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filestatsreceiver"
	fluentforwardreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver"
	googlecloudpubsubreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudpubsubreceiver"
	haproxyreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/haproxyreceiver"
	hostmetricsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	httpcheckreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/httpcheckreceiver"
	icmpcheckreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/icmpcheckreceiver"
	iisreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/iisreceiver"
	influxdbreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/influxdbreceiver"
	jaegerreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver"
	journaldreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver"
	k8sclusterreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"
	k8seventsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver"
	k8sobjectsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver"
	kafkametricsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver"
	kafkareceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver"
	kubeletstatsreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver"
	memcachedreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/memcachedreceiver"
	mongodbatlasreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver"
	mongodbreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver"
	mysqlreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver"
	nginxreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver"
	ntpreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/ntpreceiver"
	oracledbreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/oracledbreceiver"
	postgresqlreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver"
	prometheusreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	prometheusremotewritereceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusremotewritereceiver"
	purefareceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/purefareceiver"
	rabbitmqreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/rabbitmqreceiver"
	receivercreator "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/receivercreator"
	redisreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver"
	saphanareceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/saphanareceiver"
	simpleprometheusreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/simpleprometheusreceiver"
	snmpreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/snmpreceiver"
	snowflakereceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/snowflakereceiver"
	solacereceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/solacereceiver"
	splunkenterprisereceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkenterprisereceiver"
	splunkhecreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/splunkhecreceiver"
	sqlqueryreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlqueryreceiver"
	sqlserverreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlserverreceiver"
	sshcheckreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sshcheckreceiver"
	statsdreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver"
	syslogreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver"
	systemdreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/systemdreceiver"
	tcpcheckreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcpcheckreceiver"
	tcplogreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tcplogreceiver"
	tlscheckreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/tlscheckreceiver"
	udplogreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/udplogreceiver"
	vcenterreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/vcenterreceiver"
	wavefrontreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/wavefrontreceiver"
	windowseventlogreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver"
	windowsperfcountersreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver"
	windowsservicereceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsservicereceiver"
	yanggrpcreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/yanggrpcreceiver"
	zipkinreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver"
	zookeeperreceiver "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zookeeperreceiver"
	forwardconnector "go.opentelemetry.io/collector/connector/forwardconnector"
	debugexporter "go.opentelemetry.io/collector/exporter/debugexporter"
	nopexporter "go.opentelemetry.io/collector/exporter/nopexporter"
	otlpexporter "go.opentelemetry.io/collector/exporter/otlpexporter"
	otlphttpexporter "go.opentelemetry.io/collector/exporter/otlphttpexporter"
	zpagesextension "go.opentelemetry.io/collector/extension/zpagesextension"
	batchprocessor "go.opentelemetry.io/collector/processor/batchprocessor"
	memorylimiterprocessor "go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	nopreceiver "go.opentelemetry.io/collector/receiver/nopreceiver"
	otlpreceiver "go.opentelemetry.io/collector/receiver/otlpreceiver"
)

// NewBaseline returns the shared baseline component set: upstream collector
// core and contrib components only, with no Splunk-specific components. It is
// the common denominator every flavor layers onto and is not shipped on its own.
func NewBaseline() *Baseline {
	return &Baseline{
		extensions: []extension.Factory{
			ackextension.NewFactory(),
			basicauthextension.NewFactory(),
			bearertokenauthextension.NewFactory(),
			awsiamdbauthextension.NewFactory(),
			dockerobserver.NewFactory(),
			ecsobserver.NewFactory(),
			filestorage.NewFactory(),
			googlecloudlogentryencodingextension.NewFactory(),
			headerssetterextension.NewFactory(),
			healthcheckextension.NewFactory(),
			hostobserver.NewFactory(),
			httpforwarderextension.NewFactory(),
			k8sleaderelector.NewFactory(),
			k8sobserver.NewFactory(),
			oauth2clientauthextension.NewFactory(),
			opampextension.NewFactory(),
			pprofextension.NewFactory(),
			textencodingextension.NewFactory(),
			zpagesextension.NewFactory(),
		},
		receivers: []receiver.Factory{
			activedirectorydsreceiver.NewFactory(),
			apachereceiver.NewFactory(),
			apachesparkreceiver.NewFactory(),
			awscloudwatchreceiver.NewFactory(),
			awscontainerinsightreceiver.NewFactory(),
			awsecscontainermetricsreceiver.NewFactory(),
			azureblobreceiver.NewFactory(),
			azureeventhubreceiver.NewFactory(),
			azuremonitorreceiver.NewFactory(),
			carbonreceiver.NewFactory(),
			chronyreceiver.NewFactory(),
			ciscoosreceiver.NewFactory(),
			cloudfoundryreceiver.NewFactory(),
			collectdreceiver.NewFactory(),
			dnscheckreceiver.NewFactory(),
			dockerstatsreceiver.NewFactory(),
			elasticsearchreceiver.NewFactory(),
			filelogreceiver.NewFactory(),
			filestatsreceiver.NewFactory(),
			fluentforwardreceiver.NewFactory(),
			googlecloudpubsubreceiver.NewFactory(),
			haproxyreceiver.NewFactory(),
			hostmetricsreceiver.NewFactory(),
			httpcheckreceiver.NewFactory(),
			icmpcheckreceiver.NewFactory(),
			iisreceiver.NewFactory(),
			influxdbreceiver.NewFactory(),
			jaegerreceiver.NewFactory(),
			journaldreceiver.NewFactory(),
			k8sclusterreceiver.NewFactory(),
			k8seventsreceiver.NewFactory(),
			k8sobjectsreceiver.NewFactory(),
			kafkametricsreceiver.NewFactory(),
			kafkareceiver.NewFactory(),
			kubeletstatsreceiver.NewFactory(),
			memcachedreceiver.NewFactory(),
			mongodbatlasreceiver.NewFactory(),
			mongodbreceiver.NewFactory(),
			mysqlreceiver.NewFactory(),
			nginxreceiver.NewFactory(),
			nopreceiver.NewFactory(),
			ntpreceiver.NewFactory(),
			oracledbreceiver.NewFactory(),
			otlpreceiver.NewFactory(),
			postgresqlreceiver.NewFactory(),
			prometheusreceiver.NewFactory(),
			prometheusremotewritereceiver.NewFactory(),
			purefareceiver.NewFactory(),
			rabbitmqreceiver.NewFactory(),
			receivercreator.NewFactory(),
			redisreceiver.NewFactory(),
			saphanareceiver.NewFactory(),
			simpleprometheusreceiver.NewFactory(),
			snmpreceiver.NewFactory(),
			snowflakereceiver.NewFactory(),
			solacereceiver.NewFactory(),
			splunkenterprisereceiver.NewFactory(),
			splunkhecreceiver.NewFactory(),
			sqlqueryreceiver.NewFactory(),
			sqlserverreceiver.NewFactory(),
			sshcheckreceiver.NewFactory(),
			statsdreceiver.NewFactory(),
			syslogreceiver.NewFactory(),
			systemdreceiver.NewFactory(),
			tcpcheckreceiver.NewFactory(),
			tcplogreceiver.NewFactory(),
			tlscheckreceiver.NewFactory(),
			udplogreceiver.NewFactory(),
			vcenterreceiver.NewFactory(),
			wavefrontreceiver.NewFactory(),
			windowseventlogreceiver.NewFactory(),
			windowsperfcountersreceiver.NewFactory(),
			windowsservicereceiver.NewFactory(),
			yanggrpcreceiver.NewFactory(),
			zipkinreceiver.NewFactory(),
			zookeeperreceiver.NewFactory(),
		},
		processors: []processor.Factory{
			attributesprocessor.NewFactory(),
			batchprocessor.NewFactory(),
			cumulativetodeltaprocessor.NewFactory(),
			filterprocessor.NewFactory(),
			groupbyattrsprocessor.NewFactory(),
			k8sattributesprocessor.NewFactory(),
			logstransformprocessor.NewFactory(),
			memorylimiterprocessor.NewFactory(),
			metricsgenerationprocessor.NewFactory(),
			metricstransformprocessor.NewFactory(),
			probabilisticsamplerprocessor.NewFactory(),
			redactionprocessor.NewFactory(),
			resourcedetectionprocessor.NewFactory(),
			resourceprocessor.NewFactory(),
			spanprocessor.NewFactory(),
			tailsamplingprocessor.NewFactory(),
			transformprocessor.NewFactory(),
		},
		exporters: []exporter.Factory{
			awss3exporter.NewFactory(),
			debugexporter.NewFactory(),
			fileexporter.NewFactory(),
			googlecloudstorageexporter.NewFactory(),
			kafkaexporter.NewFactory(),
			loadbalancingexporter.NewFactory(),
			nopexporter.NewFactory(),
			otlpexporter.NewFactory(),
			otlphttpexporter.NewFactory(),
			prometheusremotewriteexporter.NewFactory(),
			signalfxexporter.NewFactory(),
			splunkhecexporter.NewFactory(),
		},
		connectors: []connector.Factory{
			countconnector.NewFactory(),
			forwardconnector.NewFactory(),
			routingconnector.NewFactory(),
			spanmetricsconnector.NewFactory(),
			sumconnector.NewFactory(),
		},
	}
}
