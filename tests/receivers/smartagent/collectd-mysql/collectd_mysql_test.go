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

//go:build integration

package tests

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

const (
	user     = "testuser"
	password = "testpass"
	dbName   = "testdb"
)

var mysqlServer = testutils.NewContainer().WithImage("mysql:latest").WithEnv(
	map[string]string{
		"MYSQL_DATABASE":      dbName,
		"MYSQL_USER":          user,
		"MYSQL_PASSWORD":      password,
		"MYSQL_ROOT_PASSWORD": password,
	}).WithExposedPorts("3306:3306").WithName("mysql-server").WillWaitForPorts("3306").WillWaitForLogs(
	"MySQL init process done. Ready for start up.",
	"ready for connections. Bind-address:",
)

func TestCollectdMySQLProvidesAllMetrics(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.PrintLogsOnFailure()
	defer tc.ShutdownOTLPReceiverSink()

	cntrs, shutdown := tc.Containers(mysqlServer)
	defer shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rc, r, err := cntrs[0].Exec(ctx, []string{"mysql", "-uroot", "-ptestpass", "-e", "grant PROCESS on *.* TO 'testuser'@'%'; flush privileges;"})
	if r != nil {
		defer func() {
			if t.Failed() {
				out, readErr := io.ReadAll(r)
				require.NoError(t, readErr)
				fmt.Printf("mysql:\n%s\n", string(out))
			}
		}()
	}
	require.NoError(t, err)
	require.Zero(t, rc)

	cfg := mysql.Config{
		User:   user,
		Passwd: password,
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: dbName,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	// exercise the target server for use-based metric generation
	_, err = db.Exec("CREATE TABLE a_table (name VARCHAR(255), preference VARCHAR(255))")
	require.NoError(t, err)
	_, err = db.Exec("ALTER TABLE a_table ADD COLUMN id INT AUTO_INCREMENT PRIMARY KEY")
	require.NoError(t, err)
	insert := "INSERT INTO a_table (name, preference) VALUES (?, ?)"
	_, err = db.Exec(insert, "some.name", "some preference")
	require.NoError(t, err)
	_, err = db.Exec(insert, "another.name", "another preference")
	require.NoError(t, err)
	_, err = db.Exec("UPDATE a_table SET preference = 'the real preference' WHERE name = 'some.name'")
	require.NoError(t, err)
	rows, err := db.Query("SELECT * FROM a_table")
	defer rows.Close()
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM a_table WHERE name = 'another.name'")
	require.NoError(t, err)

	testutils.AssertAllMetricsReceived(t, "all.yaml", "isolated_config.yaml", nil, nil)
}

func TestCollectdIsolatedLogger(t *testing.T) {
	tc := testutils.NewTestcase(t)
	defer tc.ShutdownOTLPReceiverSink()

	_, shutdown := tc.Containers(mysqlServer)
	defer shutdown()

	for _, test := range []struct {
		config               string
		expectedLogContent   map[string]bool
		unexpectedLogContent map[string]bool
	}{
		{
			config: "isolated_config.yaml",
			expectedLogContent: map[string]bool{
				`"collectdInstance": "monitor-smartagentcollectdmysql", "monitorID": "smartagentcollectdmysql"`: false,
				`"monitorType": "collectd/mysql"`:     false,
				`"name": "smartagent/collectd/mysql"`: false,
				`mysql plugin: Failed to store query result: Access denied; you need (at least one of) the PROCESS privilege(s) for this operation	{"kind": "receiver", "name": "smartagent/collectd/mysql", "data_type": "metrics", "collectdInstance": "monitor-smartagentcollectdmysql", "monitorID": "smartagentcollectdmysql", "monitorType": "collectd/mysql"`: false,
				`starting isolated configd instance "monitor-smartagentcollectdmysql"`: false,
			},
			unexpectedLogContent: map[string]bool{
				`"collectdInstance": "global", "monitorID": "smartagentcollectdmysql"`: false,
			},
		},
		{
			config: "not_isolated_config.yaml",
			expectedLogContent: map[string]bool{
				`mysql plugin: Failed to store query result: Access denied; you need (at least one of) the PROCESS privilege(s) for this operation	{"kind": "receiver", "name": "smartagent/collectd/mysql", "data_type": "metrics", "name": "default", "collectdInstance": "global"}`: false,
			},
			unexpectedLogContent: map[string]bool{
				`starting isolated configd instance`:                    false,
				`"collectdInstance": "monitor-smartagentcollectdmysql"`: false,
			},
		},
	} {
		t.Run(test.config, func(t *testing.T) {
			expectedContent := test.expectedLogContent
			unexpectedContent := test.unexpectedLogContent
			core, observed := observer.New(zap.DebugLevel)
			t.Cleanup(func() {
				if t.Failed() {
					fmt.Printf("Logs: \n")
					for _, statement := range observed.All() {
						fmt.Printf("%v\n", statement)
					}
				}
			})

			_, shutdownCollector := tc.SplunkOtelCollector(
				test.config, func(collector testutils.Collector) testutils.Collector {
					return collector.WithLogger(zap.New(core))
				})
			defer shutdownCollector()

			require.Eventually(t, func() bool {
				for _, l := range observed.All() {
					for expected := range expectedContent {
						if strings.Contains(l.Message, expected) {
							expectedContent[expected] = true
						}
					}
					for unexpected := range unexpectedContent {
						if strings.Contains(l.Message, unexpected) {
							unexpectedContent[unexpected] = true
						}
					}
				}
				for _, found := range expectedContent {
					if !found {
						return false
					}
				}
				for _, found := range unexpectedContent {
					if found {
						return false
					}
				}
				return true
			}, time.Minute, time.Second, "expected: %v, unexpected: %v", expectedContent, unexpectedContent)
		})
	}
}
