//go:build linux
// +build linux

package core

// Do an import of all of the built-in observers and monitors that are
// not appropriate for windows until we get a proper plugin system

import (
	// Import everything that isn't referenced anywhere else
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/cgroups"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/activemq"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/apache"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/cassandra"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/chrony"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/cpu"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/cpufreq"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/custom"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/df"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/disk"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/genericjmx"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/hadoopjmx"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/kafka"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/kafkaconsumer"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/kafkaproducer"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/load"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/memcached"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/memory"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/metadata"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/mysql"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/netinterface"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/nginx"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/postgresql"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/processes"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/protocols"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/python"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/rabbitmq"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/redis"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/solr"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/spark"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/statsd"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/tomcat"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/uptime"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/vmem"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/process"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/varnish"
)
