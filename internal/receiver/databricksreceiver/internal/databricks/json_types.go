// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package databricks

import (
	"time"

	"github.com/signalfx/splunk-otel-collector/internal/receiver/databricksreceiver/internal/spark"
)

// This file defines types into which JSON responses from the Databricks and
// Spark APIs are unmarshalled.

// jobsList is a top level type
type jobsList struct {
	Jobs    []Job `json:"jobs"`
	HasMore bool  `json:"has_more"`
}

type Job struct {
	CreatorUserName string      `json:"creator_user_name"`
	Settings        jobSettings `json:"settings"`
	JobID           int         `json:"job_id"`
	CreatedTime     int64       `json:"created_time"`
}

type jobSettings struct {
	Schedule           schedule           `json:"schedule,omitempty"`
	Name               string             `json:"name"`
	Format             string             `json:"format"`
	Tasks              []jobTask          `json:"tasks"`
	EmailNotifications emailNotifications `json:"email_notifications"`
	TimeoutSeconds     int                `json:"timeout_seconds,omitempty"`
	MaxConcurrentRuns  int                `json:"max_concurrent_runs"`
}

type emailNotifications struct {
	OnFailure             []string `json:"on_failure,omitempty"`
	OnSuccess             []string `json:"on_success,omitempty"`
	NoAlertForSkippedRuns bool     `json:"no_alert_for_skipped_runs"`
}

type jobTask struct {
	EmailNotifications struct{}     `json:"email_notifications,omitempty"`
	PythonWheelTask    any          `json:"python_wheel_task,omitempty"`
	PipelineTask       any          `json:"pipeline_task,omitempty"`
	SparkPythonTask    any          `json:"spark_python_task,omitempty"`
	SparkJarTask       any          `json:"spark_jar_task,omitempty"`
	NotebookTask       any          `json:"notebook_task,omitempty"`
	SparkSubmitTask    any          `json:"spark_submit_task,omitempty"`
	Description        string       `json:"description,omitempty"`
	ExistingClusterID  string       `json:"existing_cluster_id,omitempty"`
	TaskKey            string       `json:"task_key"`
	DependsOn          []dependency `json:"depends_on,omitempty"`
	NewCluster         newCluster   `json:"new_cluster,omitempty"`
	TimeoutSeconds     int          `json:"timeout_seconds"`
}

type dependency struct {
	TaskKey string `json:"task_key"`
}

type newCluster struct {
	ClusterName       string          `json:"cluster_name"`
	SparkVersion      string          `json:"spark_version"`
	NodeTypeID        string          `json:"node_type_id"`
	SparkEnvVars      sparkEnvVars    `json:"spark_env_vars"`
	AzureAttributes   azureAttributes `json:"azure_attributes"`
	NumWorkers        int             `json:"num_workers"`
	EnableElasticDisk bool            `json:"enable_elastic_disk"`
}

type sparkEnvVars struct {
	PysparkPython string `json:"PYSPARK_PYTHON"`
}

type azureAttributes struct {
	Availability    string  `json:"availability"`
	FirstOnDemand   int     `json:"first_on_demand"`
	SpotBidMaxPrice float64 `json:"spot_bid_max_price"`
}

// jobRuns is a top-level type
type jobRuns struct {
	Runs    []JobRun `json:"runs"`
	HasMore bool     `json:"has_more"`
}

type JobRun struct {
	Schedule             schedule     `json:"schedule"`
	Message              string       `json:"message"`
	Format               string       `json:"format"`
	CreatorUserName      string       `json:"creator_user_name"`
	RunType              string       `json:"run_type"`
	RunPageURL           string       `json:"run_page_url"`
	RunName              string       `json:"run_name"`
	Trigger              string       `json:"trigger"`
	ErrorCode            string       `json:"error_code"`
	State                state        `json:"state"`
	Tasks                []jobRunTask `json:"tasks"`
	CleanupDuration      int          `json:"cleanup_duration"`
	ExecutionDuration    int          `json:"execution_duration"`
	StartTime            int64        `json:"start_time"`
	OriginalAttemptRunID int          `json:"original_attempt_run_id"`
	NumberInJob          int          `json:"number_in_job"`
	RunID                int          `json:"run_id"`
	JobID                int          `json:"job_id"`
	SetupDuration        int          `json:"setup_duration"`
	EndTime              int64        `json:"end_time"`
}

type schedule struct {
	QuartzCronExpression string `json:"quartz_cron_expression"`
	TimezoneID           string `json:"timezone_id"`
	PauseStatus          string `json:"pause_status"`
}

type jobRunTask struct {
	ClusterInstance   clusterInstance `json:"cluster_instance"`
	TaskKey           string          `json:"task_key"`
	NotebookTask      notebookTask    `json:"notebook_task"`
	ExistingClusterID string          `json:"existing_cluster_id"`
	RunPageURL        string          `json:"run_page_url"`
	State             state           `json:"state"`
	RunID             int             `json:"run_id"`
	SetupDuration     int             `json:"setup_duration"`
	ExecutionDuration int             `json:"execution_duration"`
	CleanupDuration   int             `json:"cleanup_duration"`
	EndTime           int64           `json:"end_time"`
	StartTime         int64           `json:"start_time"`
	AttemptNumber     int             `json:"attempt_number"`
}

type clusterInstance struct {
	ClusterID      string `json:"cluster_id"`
	SparkContextID string `json:"spark_context_id,omitempty"`
}

type notebookTask struct {
	NotebookPath string `json:"notebook_path"`
}

type state struct {
	LifeCycleState          string `json:"life_cycle_state"`
	StateMessage            string `json:"state_message"`
	ResultState             string `json:"result_state,omitempty"`
	UserCancelledOrTimedout bool   `json:"user_cancelled_or_timedout"`
}

// clusterList is a top-level type
type clusterList struct {
	Clusters []spark.Cluster `json:"clusters"`
}

// pipelinesInfo is a top-level type
type pipelinesInfo struct {
	Statuses []struct {
		PipelineID      string `json:"pipeline_id"`
		State           string `json:"state"`
		Name            string `json:"name"`
		CreatorUserName string `json:"creator_user_name"`
		RunAsUserName   string `json:"run_as_user_name"`
		LatestUpdates   []struct {
			CreationTime time.Time `json:"creation_time"`
			UpdateID     string    `json:"update_id"`
			State        string    `json:"state"`
		} `json:"latest_updates"`
	} `json:"statuses"`
}

type pipelineInfo struct {
	PipelineID      string `json:"pipeline_id"`
	State           string `json:"state"`
	ClusterID       string `json:"cluster_id"`
	Name            string `json:"name"`
	CreatorUserName string `json:"creator_user_name"`
	RunAsUserName   string `json:"run_as_user_name"`
	LatestUpdates   []struct {
		CreationTime time.Time `json:"creation_time"`
		UpdateID     string    `json:"update_id"`
		State        string    `json:"state"`
	} `json:"latest_updates"`
	Spec struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Storage  string `json:"storage"`
		Edition  string `json:"edition"`
		Channel  string `json:"channel"`
		Clusters []struct {
			Label     string `json:"label"`
			Autoscale struct {
				Mode       string `json:"mode"`
				MinWorkers int    `json:"min_workers"`
				MaxWorkers int    `json:"max_workers"`
			} `json:"autoscale"`
		} `json:"clusters"`
		Libraries []struct {
			Notebook struct {
				Path string `json:"path"`
			} `json:"notebook"`
		} `json:"libraries"`
		Continuous  bool `json:"continuous"`
		Development bool `json:"development"`
		Photon      bool `json:"photon"`
	} `json:"spec"`
	LastModified int64 `json:"last_modified"`
}
