package hana

import (
	"fmt"

	"github.com/signalfx/signalfx-agent/pkg/monitors/sql"
)

const defaultMaxExpensiveQueries = 10

func queries(maxExpensiveQueries int) []sql.Query {
	if maxExpensiveQueries == 0 {
		maxExpensiveQueries = defaultMaxExpensiveQueries
	}
	return []sql.Query{
		{
			Query: `SELECT host AS hana_host, usage_type, used_size FROM m_disk_usage WHERE used_size >= 0;`,
			Metrics: []sql.Metric{
				{
					MetricName:       "sap.hana.disk.used_size",
					ValueColumn:      "used_size",
					DimensionColumns: []string{"hana_host", "usage_type"},
				},
			},
		},
		{
			Query: `SELECT host AS hana_host, SUM(total_device_size) AS total_size FROM (SELECT device_id, host, MAX(total_device_size) AS total_device_size FROM m_disks GROUP BY device_id, host) GROUP BY host;`,
			Metrics: []sql.Metric{
				{
					MetricName:       "sap.hana.disk.total_size",
					ValueColumn:      "total_size",
					DimensionColumns: []string{"hana_host"},
				},
			},
		},
		{
			/*
				TODO filter out <0 rows?

				HANA_HOST,SERVICE_NAME,PROCESS_CPU,OPEN_FILE_COUNT
				"abc123","daemon",-1,-1
				"abc123","nameserver",0,71
				"abc123","compileserver",0,15
				"abc123","indexserver",0,91
				"abc123","dpserver",0,61
				"abc123","diserver",0,16
			*/
			Query: `SELECT host AS hana_host, service_name, process_cpu, open_file_count FROM m_service_statistics;`,
			Metrics: []sql.Metric{
				{
					MetricName:       "sap.hana.service.cpu.utilization",
					ValueColumn:      "process_cpu",
					DimensionColumns: []string{"hana_host", "service_name"},
				},
				{
					MetricName:       "sap.hana.service.file.open",
					ValueColumn:      "open_file_count",
					DimensionColumns: []string{"hana_host", "service_name"},
				},
			},
		},
		{
			/*
				HANA_HOST,FREE_PHYSICAL_MEMORY,USED_PHYSICAL_MEMORY,FREE_SWAP_SPACE,USED_SWAP_SPACE,ALLOCATION_LIMIT,INSTANCE_TOTAL_MEMORY_USED_SIZE,INSTANCE_TOTAL_MEMORY_ALLOCATED_SIZE,INSTANCE_CODE_SIZE,INSTANCE_SHARED_MEMORY_ALLOCATED_SIZE,OPEN_FILE_COUNT,TOTAL_CPU_USER_TIME,TOTAL_CPU_SYSTEM_TIME,TOTAL_CPU_WIO_TIME,TOTAL_CPU_IDLE_TIME
				"abc123",26931253248,4966428672,0,0,28707913728,5357671063,8904347648,2464563200,5627904,11392,393490,90850,0,873437820
			*/
			Query: `SELECT host AS hana_host, free_physical_memory, used_physical_memory, free_swap_space, used_swap_space, allocation_limit, instance_total_memory_used_size, instance_total_memory_allocated_size, instance_code_size, instance_shared_memory_allocated_size, open_file_count, total_cpu_user_time, total_cpu_system_time, total_cpu_wio_time, total_cpu_idle_time FROM m_host_resource_utilization;`,
			Metrics: []sql.Metric{
				{
					MetricName:       "sap.hana.host.memory.physical.free",
					ValueColumn:      "free_physical_memory",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.host.memory.physical.used",
					ValueColumn:      "used_physical_memory",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.host.memory.swap.free",
					ValueColumn:      "free_swap_space",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.host.memory.swap.used",
					ValueColumn:      "used_swap_space",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.host.memory.allocation_limit",
					ValueColumn:      "allocation_limit",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.host.memory.total_used",
					ValueColumn:      "instance_total_memory_used_size",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.host.memory.total_allocated",
					ValueColumn:      "instance_total_memory_allocated_size",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.host.memory.code",
					ValueColumn:      "instance_code_size",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.host.memory.shared",
					ValueColumn:      "instance_shared_memory_allocated_size",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.host.file.open",
					ValueColumn:      "open_file_count",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.host.cpu.user",
					ValueColumn:      "total_cpu_user_time",
					DimensionColumns: []string{"hana_host"},
					IsCumulative:     true,
				},
				{
					MetricName:       "sap.hana.host.cpu.system",
					ValueColumn:      "total_cpu_system_time",
					DimensionColumns: []string{"hana_host"},
					IsCumulative:     true,
				},
				{
					MetricName:       "sap.hana.host.cpu.wio",
					ValueColumn:      "total_cpu_wio_time",
					DimensionColumns: []string{"hana_host"},
					IsCumulative:     true,
				},
				{
					MetricName:       "sap.hana.host.cpu.idle",
					ValueColumn:      "total_cpu_idle_time",
					DimensionColumns: []string{"hana_host"},
					IsCumulative:     true,
				},
			},
		},
		{
			/*
				HANA_HOST,SERVICE_NAME,LOGICAL_MEMORY_SIZE,PHYSICAL_MEMORY_SIZE,CODE_SIZE,STACK_SIZE,HEAP_MEMORY_ALLOCATED_SIZE,HEAP_MEMORY_USED_SIZE,SHARED_MEMORY_ALLOCATED_SIZE,SHARED_MEMORY_USED_SIZE,ALLOCATION_LIMIT,EFFECTIVE_ALLOCATION_LIMIT,TOTAL_MEMORY_USED_SIZE
				"abc123","nameserver",4416409600,1495285760,2421600256,98439168,997945344,742072423,0,0,28707913728,26489645385,3163672679
				"abc123","compileserver",1009868800,309551104,323026944,37044224,652439552,65553792,0,0,28707913728,23576112226,388580736
				"abc123","indexserver",6177243136,3288928256,2420973568,220229632,2752126976,1379229923,0,0,28707913728,26991900813,3800203491
				"abc123","dpserver",4951605248,1662558208,2433900544,125722624,1448525824,635247555,0,0,28707913728,26260884789,3069148099
				"abc123","diserver",2449539072,528445440,1715077120,35721216,687476736,94015225,0,0,28707913728,25000828379,1809092345
			*/
			Query: `SELECT host AS hana_host, service_name, logical_memory_size, physical_memory_size, code_size, stack_size, heap_memory_allocated_size, heap_memory_used_size, shared_memory_allocated_size, shared_memory_used_size, allocation_limit, effective_allocation_limit, total_memory_used_size FROM m_service_memory;`,
			Metrics: []sql.Metric{
				{
					MetricName:       "sap.hana.service.memory.logical",
					ValueColumn:      "logical_memory_size",
					DimensionColumns: []string{"hana_host", "service_name"},
				},
				{
					MetricName:       "sap.hana.service.memory.physical",
					ValueColumn:      "physical_memory_size",
					DimensionColumns: []string{"hana_host", "service_name"},
				},
				{
					MetricName:       "sap.hana.service.memory.code",
					ValueColumn:      "code_size",
					DimensionColumns: []string{"hana_host", "service_name"},
				},
				{
					MetricName:       "sap.hana.service.memory.stack",
					ValueColumn:      "stack_size",
					DimensionColumns: []string{"hana_host", "service_name"},
				},
				{
					MetricName:       "sap.hana.service.memory.heap.allocated",
					ValueColumn:      "heap_memory_allocated_size",
					DimensionColumns: []string{"hana_host", "service_name"},
				},
				{
					MetricName:       "sap.hana.service.memory.heap.used",
					ValueColumn:      "heap_memory_used_size",
					DimensionColumns: []string{"hana_host", "service_name"},
				},
				{
					MetricName:       "sap.hana.service.memory.shared.allocated",
					ValueColumn:      "shared_memory_allocated_size",
					DimensionColumns: []string{"hana_host", "service_name"},
				},
				{
					MetricName:       "sap.hana.service.memory.shared.used",
					ValueColumn:      "shared_memory_used_size",
					DimensionColumns: []string{"hana_host", "service_name"},
				},
				{
					MetricName:       "sap.hana.service.memory.allocation_limit",
					ValueColumn:      "allocation_limit",
					DimensionColumns: []string{"hana_host", "service_name"},
				},
				{
					MetricName:       "sap.hana.service.memory.allocation_limit_effective",
					ValueColumn:      "effective_allocation_limit",
					DimensionColumns: []string{"hana_host", "service_name"},
				},
				{
					MetricName:       "sap.hana.service.memory.total_used",
					ValueColumn:      "total_memory_used_size",
					DimensionColumns: []string{"hana_host", "service_name"},
				},
			},
		},
		{
			/*
				HANA_HOST,SERVICE_NAME,COMPONENT_NAME,USED_MEMORY_SIZE
				"abc123","indexserver","System",506506042
				"abc123","indexserver","Monitoring & Statistical Data",171169440
				"abc123","indexserver","Statement Execution & Intermediate Results",310607365
				"abc123","indexserver","Caches",187397869
				"abc123","indexserver","Column Store Tables",55257512
				"abc123","dpserver","System",338552595
				"abc123","dpserver","Monitoring & Statistical Data",76194080
				"abc123","dpserver","Statement Execution & Intermediate Results",149375312
				"abc123","diserver","System",37725617
				"abc123","dpserver","Caches",2990112
				"abc123","diserver","Monitoring & Statistical Data",15975616
				"abc123","diserver","Column Store Tables",10928
				"abc123","diserver","Statement Execution & Intermediate Results",40363936
				"abc123","diserver","Basis System",80
				"abc123","dpserver","Column Store Tables",11656
				"abc123","indexserver","Basis System",80
				"abc123","diserver","Caches",53576
				"abc123","diserver","Other",0
				"abc123","dpserver","Basis System",80
				"abc123","indexserver","Other",0
				"abc123","dpserver","Other",0
				"abc123","indexserver","Row Store Tables",153116672
				"abc123","dpserver","Row Store Tables",67137728
				"abc123","nameserver","Code Size",2421600256
				"abc123","compileserver","Code Size",323026944
				"abc123","indexserver","Code Size",2420973568
				"abc123","dpserver","Code Size",2433900544
				"abc123","diserver","Code Size",1715077120
				"abc123","nameserver","Stack Size",98439168
				"abc123","compileserver","Stack Size",37044224
				"abc123","indexserver","Stack Size",221552640
				"abc123","dpserver","Stack Size",125722624
				"abc123","diserver","Stack Size",35721216
			*/
			Query: `SELECT services.host AS hana_host, services.service_name AS service_name, memory.component AS component_name, memory.used_memory_size AS used_memory_size FROM m_service_component_memory AS memory JOIN m_services AS services ON memory.host = services.host AND memory.port = services.port;`,
			Metrics: []sql.Metric{
				{
					MetricName:       "sap.hana.service.component.memory.used",
					ValueColumn:      "used_memory_size",
					DimensionColumns: []string{"hana_host", "service_name", "component_name"},
				},
			},
		},
		{
			/*
				HANA_HOST,STATEMENT_COUNT,AVG_EXECUTION_TIME,MAX_EXECUTION_TIME,TOTAL_EXECUTION_TIME,AVG_EXECUTION_MEMORY_SIZE,MAX_EXECUTION_MEMORY_SIZE,TOTAL_EXECUTION_MEMORY_SIZE
				"abc123",1,0,0,0,0,0,0
			*/
			Query: `SELECT host AS hana_host, COUNT(*) AS statement_count, TO_DOUBLE(AVG(avg_execution_time)) AS avg_execution_time, MAX(max_execution_time) AS max_execution_time, SUM(total_execution_time) AS total_execution_time, TO_DOUBLE(AVG(avg_execution_memory_size)) AS avg_execution_memory_size, MAX(max_execution_memory_size) AS max_execution_memory_size, SUM(total_execution_memory_size) AS total_execution_memory_size FROM m_active_statements GROUP BY host;`,
			Metrics: []sql.Metric{
				{
					MetricName:       "sap.hana.statement.active.count",
					ValueColumn:      "statement_count",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.statement.active.execution.time.mean",
					ValueColumn:      "avg_execution_time",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.statement.active.execution.time.sum",
					ValueColumn:      "total_execution_time",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.statement.active.execution.time.max",
					ValueColumn:      "max_execution_time",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.statement.active.execution.memory.mean",
					ValueColumn:      "avg_execution_memory_size",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.statement.active.execution.memory.max",
					ValueColumn:      "max_execution_memory_size",
					DimensionColumns: []string{"hana_host"},
				},
				{
					MetricName:       "sap.hana.statement.active.execution.memory.sum",
					ValueColumn:      "total_execution_memory_size",
					DimensionColumns: []string{"hana_host"},
				},
			},
		},
		{
			/*
				0 rows selected (overall time 70.721 msec; server time 3765 usec)
			*/
			Query: fmt.Sprintf(`SELECT host AS hana_host,statement_hash,db_user,schema_name,app_user,operation,SUM(duration_microsec) AS total_duration_microsec,SUM(records) AS total_records,SUM(cpu_time) AS total_cpu_time,SUM(lock_wait_duration) AS total_lock_wait_duration,COUNT(*) AS count FROM m_expensive_statements WHERE start_time > ADD_DAYS(CURRENT_TIMESTAMP , -1) AND (host , schema_name , statement_hash) IN ( SELECT host , schema_name , statement_hash FROM (SELECT * , rank() OVER (PARTITION BY host , schema_name ORDER BY duration_microsec DESC) AS rank FROM (SELECT host , schema_name , statement_hash , MAX (duration_microsec) AS duration_microsec FROM m_expensive_statements WHERE start_time > ADD_DAYS(CURRENT_TIMESTAMP , -1) GROUP BY host , schema_name , statement_hash)) WHERE rank < %d )GROUP BY host, statement_hash, db_user, schema_name, app_user, operation;`, maxExpensiveQueries),
			Metrics: []sql.Metric{
				{
					MetricName:       "sap.hana.statement.expensive.count",
					ValueColumn:      "count",
					DimensionColumns: []string{"hana_host", "statement_hash", "db_user", "schema_name", "app_user", "operation"},
					IsCumulative:     true,
				},
				{
					MetricName:       "sap.hana.statement.expensive.duration",
					ValueColumn:      "total_duration_microsec",
					DimensionColumns: []string{"hana_host", "statement_hash", "db_user", "schema_name", "app_user", "operation"},
				},
				{
					MetricName:       "sap.hana.statement.expensive.records",
					ValueColumn:      "total_records",
					DimensionColumns: []string{"hana_host", "statement_hash", "db_user", "schema_name", "app_user", "operation"},
					IsCumulative:     true,
				},
				{
					MetricName:       "sap.hana.statement.expensive.cpu_time",
					ValueColumn:      "total_cpu_time",
					DimensionColumns: []string{"hana_host", "statement_hash", "db_user", "schema_name", "app_user", "operation"},
					IsCumulative:     true,
				},
				{
					MetricName:       "sap.hana.statement.expensive.lock_wait_duration",
					ValueColumn:      "total_lock_wait_duration",
					DimensionColumns: []string{"hana_host", "statement_hash", "db_user", "schema_name", "app_user", "operation"},
					IsCumulative:     true,
				},
			},
		},
		{
			/*
				0 rows selected (overall time 80.065 msec; server time 3311 usec)
			*/
			Query: `SELECT host AS hana_host, statement_hash, db_user, schema_name, app_user, operation, COUNT(*) AS errors FROM m_expensive_statements WHERE error_code <> 0 GROUP BY host, statement_hash, db_user, schema_name, app_user, operation;`,
			Metrics: []sql.Metric{
				{
					MetricName:       "sap.hana.statement.expensive.errors",
					ValueColumn:      "errors",
					DimensionColumns: []string{"hana_host", "statement_hash", "db_user", "schema_name", "app_user", "operation"},
					IsCumulative:     true,
				},
			},
		},
		{
			/*
				HANA_HOST,CONNECTION_STATUS,COUNT,FETCHED_RECORD_COUNT,AFFECTED_RECORD_COUNT,SENT_MESSAGE_SIZE,SENT_MESSAGE_COUNT,RECEIVED_MESSAGE_SIZE,RECEIVED_MESSAGE_COUNT
				"abc123","IDLE",16,164630,81242,860670,683,166224,683
				"abc123","RUNNING",1,48,0,14726,25,11104,26
			*/
			Query: `SELECT host AS hana_host, connection_status, COUNT(*) AS count, SUM(fetched_record_count) AS fetched_record_count, SUM(affected_record_count) AS affected_record_count, SUM(sent_message_size) AS sent_message_size, SUM(sent_message_count) AS sent_message_count, SUM(received_message_size) AS received_message_size, SUM(received_message_count) AS received_message_count FROM m_connections GROUP BY host, connection_status HAVING connection_status != '';`,
			Metrics: []sql.Metric{
				{
					MetricName:       "sap.hana.connection.count",
					ValueColumn:      "count",
					DimensionColumns: []string{"hana_host", "connection_status"},
				},
				{
					MetricName:       "sap.hana.connection.record.fetched",
					ValueColumn:      "fetched_record_count",
					DimensionColumns: []string{"hana_host", "connection_status"},
				},
				{
					MetricName:       "sap.hana.connection.record.affected",
					ValueColumn:      "affected_record_count",
					DimensionColumns: []string{"hana_host", "connection_status"},
				},
				{
					MetricName:       "sap.hana.connection.message.sent.size",
					ValueColumn:      "sent_message_size",
					DimensionColumns: []string{"hana_host", "connection_status"},
				},
				{
					MetricName:       "sap.hana.connection.message.sent.count",
					ValueColumn:      "sent_message_count",
					DimensionColumns: []string{"hana_host", "connection_status"},
				},
				{
					MetricName:       "sap.hana.connection.message.received.size",
					ValueColumn:      "received_message_size",
					DimensionColumns: []string{"hana_host", "connection_status"},
				},
				{
					MetricName:       "sap.hana.connection.message.received.count",
					ValueColumn:      "received_message_count",
					DimensionColumns: []string{"hana_host", "connection_status"},
				},
			},
		},
		{
			/*
				HANA_HOST,TYPE,TOTAL_READS,TOTAL_TRIGGER_ASYNC_READS,TOTAL_FAILED_READS,TOTAL_READ_SIZE,TOTAL_READ_TIME,TOTAL_APPENDS,TOTAL_WRITES,TOTAL_TRIGGER_ASYNC_WRITES,TOTAL_FAILED_WRITES,TOTAL_WRITE_SIZE,TOTAL_WRITE_TIME,TOTAL_IO_TIME
				"abc123","ROOTKEY_BACKUP",0,0,0,0,0,0,0,20,0,81920,23903,23903
				"abc123","LOG",28,24,0,457330688,3779809,0,22,10080,0,51634176,7831360,11610508
				"abc123","DATA",2,12433,0,814223360,10085151,0,128,22835,0,7046029312,22361122,30091845
			*/
			Query: `SELECT host AS hana_host, type, SUM(total_reads) AS total_reads, SUM(total_trigger_async_reads) AS total_trigger_async_reads, SUM(total_failed_reads) AS total_failed_reads, SUM(total_read_size) AS total_read_size, SUM(total_read_time) AS total_read_time, SUM(total_appends) AS total_appends, SUM(total_writes) AS total_writes, SUM(total_trigger_async_writes) AS total_trigger_async_writes, SUM(total_failed_writes) AS total_failed_writes, SUM(total_write_size) AS total_write_size, SUM(total_write_time) AS total_write_time, SUM(total_io_time) AS total_io_time FROM m_volume_io_total_statistics GROUP BY host, type;`,
			Metrics: []sql.Metric{
				{
					MetricName:       "sap.hana.io.read.count",
					ValueColumn:      "total_reads",
					IsCumulative:     true,
					DimensionColumns: []string{"hana_host", "type"},
				},
				{
					MetricName:       "sap.hana.io.read.async.count",
					ValueColumn:      "total_trigger_async_reads",
					IsCumulative:     true,
					DimensionColumns: []string{"hana_host", "type"},
				},
				{
					MetricName:       "sap.hana.io.read.failed",
					ValueColumn:      "total_failed_reads",
					IsCumulative:     true,
					DimensionColumns: []string{"hana_host", "type"},
				},
				{
					MetricName:       "sap.hana.io.read.size",
					ValueColumn:      "total_read_size",
					IsCumulative:     true,
					DimensionColumns: []string{"hana_host", "type"},
				},
				{
					MetricName:       "sap.hana.io.read.time",
					ValueColumn:      "total_read_time",
					IsCumulative:     true,
					DimensionColumns: []string{"hana_host", "type"},
				},
				{
					MetricName:       "sap.hana.io.append.count",
					ValueColumn:      "total_appends",
					IsCumulative:     true,
					DimensionColumns: []string{"hana_host", "type"},
				},
				{
					MetricName:       "sap.hana.io.write.count",
					ValueColumn:      "total_writes",
					IsCumulative:     true,
					DimensionColumns: []string{"hana_host", "type"},
				},
				{
					MetricName:       "sap.hana.io.write.async.count",
					ValueColumn:      "total_trigger_async_writes",
					IsCumulative:     true,
					DimensionColumns: []string{"hana_host", "type"},
				},
				{
					MetricName:       "sap.hana.io.write.failed",
					ValueColumn:      "total_failed_writes",
					IsCumulative:     true,
					DimensionColumns: []string{"hana_host", "type"},
				},
				{
					MetricName:       "sap.hana.io.write.size",
					ValueColumn:      "total_write_size",
					IsCumulative:     true,
					DimensionColumns: []string{"hana_host", "type"},
				},
				{
					MetricName:       "sap.hana.io.write.time",
					ValueColumn:      "total_write_time",
					IsCumulative:     true,
					DimensionColumns: []string{"hana_host", "type"},
				},
				{
					MetricName:       "sap.hana.io.total.time",
					ValueColumn:      "total_io_time",
					IsCumulative:     true,
					DimensionColumns: []string{"hana_host", "type"},
				},
			},
		},
		{
			/*
				SCHEMA_NAME,TABLE_NAME,TABLE_TYPE,RECORD_COUNT,TABLE_SIZE
				"PAL_STEM_SCHEMA","PAL_TEST_IN_TYPE","ROW",0,0
				"PAL_STEM_SCHEMA","AUTO_TABLE_STEM","COLUMN",0,1376
				"SYSHDL","CONTAINER_LOCK","COLUMN",0,1376
			*/
			Query: `SELECT schema_name, table_name, table_type,	record_count, table_size FROM m_tables WHERE schema_name NOT IN ('SYS', 'SAP_PA_APL', 'BROKER_PO_USER') AND schema_name NOT LIKE '_SYS_%';`,
			Metrics: []sql.Metric{
				{
					MetricName:       "sap.hana.table.record.count",
					ValueColumn:      "record_count",
					DimensionColumns: []string{"schema_name", "table_name", "table_type"},
				},
				{
					MetricName:       "sap.hana.table.size",
					ValueColumn:      "table_size",
					DimensionColumns: []string{"schema_name", "table_name", "table_type"},
				},
			},
		},
	}
}
