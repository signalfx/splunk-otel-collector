package core

// Do an import of all of the built-in observers and monitors
// that apply to all platforms until we get a proper plugin system

import (
	// Import everything that isn't referenced anywhere else
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/appmesh"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/cadvisor"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/cloudfoundry"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/consul"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/couchbase"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/hadoop"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/jenkins"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/opcache"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/openstack"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/php"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/python"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/rabbitmq"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/redis"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/spark"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/systemd"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/collectd/zookeeper"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/conviva"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/coredns"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/cpu"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/diskio"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/docker"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/ecs"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/elasticsearch/query"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/elasticsearch/stats"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/etcd"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/expvar"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/filesystems"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/forwarder"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/gitlab"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/hana"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/haproxy"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/heroku"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/http"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/internalmetrics"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/jaegergrpc"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/jmx"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/kubernetes"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/load"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/logstash/logstash"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/logstash/tcp"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/memory"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/mongodb/atlas"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/nagios"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/netio"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/ntp"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/postgresql"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/processlist"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/prometheus/go"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/prometheus/nginxingress"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/prometheus/nginxvts"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/prometheus/node"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/prometheus/postgres"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/prometheus/prometheus"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/prometheus/redis"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/prometheus/velero"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/prometheusexporter"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/sql"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/statsd"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/subproc/signalfx/java"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/subproc/signalfx/python"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/supervisor"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/dns"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/exec"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/mssqlserver"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/ntpq"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/procstat"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/tail"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/telegraflogparser"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/telegrafsnmp"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/telegrafstatsd"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/winperfcounters"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/winservices"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/traceforwarder"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/traefik"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/vmem"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/vsphere"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/windowsiis"
	_ "github.com/signalfx/signalfx-agent/pkg/monitors/windowslegacy"
)
