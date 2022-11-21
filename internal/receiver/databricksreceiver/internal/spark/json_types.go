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

package spark

type ClusterMetrics struct {
	Gauges     map[string]Gauge
	Counters   map[string]Counter
	Histograms map[string]Histogram
	Meters     map[string]any
	Timers     map[string]Timer
}

type Gauge struct {
	Value float64 `json:"value"`
}

type Counter struct {
	Count int `json:"count"`
}

type Histogram struct {
	Count  int     `json:"count"`
	Max    int     `json:"max"`
	Mean   float64 `json:"mean"`
	Min    int     `json:"min"`
	P50    float64 `json:"p50"`
	P75    float64 `json:"p75"`
	P95    float64 `json:"p95"`
	P98    float64 `json:"p98"`
	P99    float64 `json:"p99"`
	P999   float64 `json:"p999"`
	Stddev float64 `json:"stddev"`
}

type Timer struct {
	Count         int     `json:"count"`
	Max           float64 `json:"max"`
	Mean          float64 `json:"mean"`
	Min           float64 `json:"min"`
	P50           float64 `json:"p50"`
	P75           float64 `json:"p75"`
	P95           float64 `json:"p95"`
	P98           float64 `json:"p98"`
	P99           float64 `json:"p99"`
	P999          float64 `json:"p999"`
	Stddev        float64 `json:"stddev"`
	M15Rate       float64 `json:"m15_rate"`
	M1Rate        float64 `json:"m1_rate"`
	M5Rate        float64 `json:"m5_rate"`
	MeanRate      float64 `json:"mean_rate"`
	DurationUnits string  `json:"duration_units"`
	RateUnits     string  `json:"rate_units"`
}

type ExecutorInfo struct {
	Id                string `json:"id"`
	HostPort          string `json:"hostPort"`
	IsActive          bool   `json:"isActive"`
	RddBlocks         int    `json:"rddBlocks"`
	MemoryUsed        int    `json:"memoryUsed"`
	DiskUsed          int    `json:"diskUsed"`
	TotalCores        int    `json:"totalCores"`
	MaxTasks          int    `json:"maxTasks"`
	ActiveTasks       int    `json:"activeTasks"`
	FailedTasks       int    `json:"failedTasks"`
	CompletedTasks    int    `json:"completedTasks"`
	TotalTasks        int    `json:"totalTasks"`
	TotalDuration     int    `json:"totalDuration"`
	TotalGCTime       int    `json:"totalGCTime"`
	TotalInputBytes   int64  `json:"totalInputBytes"`
	TotalShuffleRead  int    `json:"totalShuffleRead"`
	TotalShuffleWrite int    `json:"totalShuffleWrite"`
	IsBlacklisted     bool   `json:"isBlacklisted"`
	MaxMemory         int64  `json:"maxMemory"`
	AddTime           string `json:"addTime"`
	ExecutorLogs      struct {
		Stdout string `json:"stdout,omitempty"`
		Stderr string `json:"stderr,omitempty"`
	} `json:"executorLogs"`
	MemoryMetrics struct {
		UsedOnHeapStorageMemory   int   `json:"usedOnHeapStorageMemory"`
		UsedOffHeapStorageMemory  int   `json:"usedOffHeapStorageMemory"`
		TotalOnHeapStorageMemory  int64 `json:"totalOnHeapStorageMemory"`
		TotalOffHeapStorageMemory int   `json:"totalOffHeapStorageMemory"`
	} `json:"memoryMetrics"`
	BlacklistedInStages []interface{} `json:"blacklistedInStages"`
	PeakMemoryMetrics   struct {
		JVMHeapMemory              int64 `json:"JVMHeapMemory"`
		JVMOffHeapMemory           int   `json:"JVMOffHeapMemory"`
		OnHeapExecutionMemory      int   `json:"OnHeapExecutionMemory"`
		OffHeapExecutionMemory     int   `json:"OffHeapExecutionMemory"`
		OnHeapStorageMemory        int   `json:"OnHeapStorageMemory"`
		OffHeapStorageMemory       int   `json:"OffHeapStorageMemory"`
		OnHeapUnifiedMemory        int   `json:"OnHeapUnifiedMemory"`
		OffHeapUnifiedMemory       int   `json:"OffHeapUnifiedMemory"`
		DirectPoolMemory           int   `json:"DirectPoolMemory"`
		MappedPoolMemory           int   `json:"MappedPoolMemory"`
		ProcessTreeJVMVMemory      int   `json:"ProcessTreeJVMVMemory"`
		ProcessTreeJVMRSSMemory    int   `json:"ProcessTreeJVMRSSMemory"`
		ProcessTreePythonVMemory   int   `json:"ProcessTreePythonVMemory"`
		ProcessTreePythonRSSMemory int   `json:"ProcessTreePythonRSSMemory"`
		ProcessTreeOtherVMemory    int   `json:"ProcessTreeOtherVMemory"`
		ProcessTreeOtherRSSMemory  int   `json:"ProcessTreeOtherRSSMemory"`
		MinorGCCount               int   `json:"MinorGCCount"`
		MinorGCTime                int   `json:"MinorGCTime"`
		MajorGCCount               int   `json:"MajorGCCount"`
		MajorGCTime                int   `json:"MajorGCTime"`
		TotalGCTime                int   `json:"TotalGCTime"`
	} `json:"peakMemoryMetrics"`
	Attributes struct {
	} `json:"attributes"`
	Resources struct {
	} `json:"resources"`
	ResourceProfileId int           `json:"resourceProfileId"`
	IsExcluded        bool          `json:"isExcluded"`
	ExcludedInStages  []interface{} `json:"excludedInStages"`
}

type Application struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type JobInfo struct {
	JobId               int    `json:"jobId"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	SubmissionTime      string `json:"submissionTime"`
	CompletionTime      string `json:"completionTime"`
	StageIds            []int  `json:"stageIds"`
	JobGroup            string `json:"jobGroup"`
	Status              string `json:"status"`
	NumTasks            int    `json:"numTasks"`
	NumActiveTasks      int    `json:"numActiveTasks"`
	NumCompletedTasks   int    `json:"numCompletedTasks"`
	NumSkippedTasks     int    `json:"numSkippedTasks"`
	NumFailedTasks      int    `json:"numFailedTasks"`
	NumKilledTasks      int    `json:"numKilledTasks"`
	NumCompletedIndices int    `json:"numCompletedIndices"`
	NumActiveStages     int    `json:"numActiveStages"`
	NumCompletedStages  int    `json:"numCompletedStages"`
	NumSkippedStages    int    `json:"numSkippedStages"`
	NumFailedStages     int    `json:"numFailedStages"`
}

type StageInfo struct {
	Status                       string `json:"status"`
	StageId                      int    `json:"stageId"`
	AttemptId                    int    `json:"attemptId"`
	NumTasks                     int    `json:"numTasks"`
	NumActiveTasks               int    `json:"numActiveTasks"`
	NumCompleteTasks             int    `json:"numCompleteTasks"`
	NumFailedTasks               int    `json:"numFailedTasks"`
	NumKilledTasks               int    `json:"numKilledTasks"`
	NumCompletedIndices          int    `json:"numCompletedIndices"`
	SubmissionTime               string `json:"submissionTime"`
	FirstTaskLaunchedTime        string `json:"firstTaskLaunchedTime"`
	CompletionTime               string `json:"completionTime"`
	ExecutorDeserializeTime      int    `json:"executorDeserializeTime"`
	ExecutorDeserializeCpuTime   int    `json:"executorDeserializeCpuTime"`
	ExecutorRunTime              int    `json:"executorRunTime"`
	ExecutorCpuTime              int    `json:"executorCpuTime"`
	ResultSize                   int    `json:"resultSize"`
	JvmGcTime                    int    `json:"jvmGcTime"`
	ResultSerializationTime      int    `json:"resultSerializationTime"`
	MemoryBytesSpilled           int    `json:"memoryBytesSpilled"`
	DiskBytesSpilled             int    `json:"diskBytesSpilled"`
	PeakExecutionMemory          int    `json:"peakExecutionMemory"`
	InputBytes                   int    `json:"inputBytes"`
	InputRecords                 int    `json:"inputRecords"`
	OutputBytes                  int    `json:"outputBytes"`
	OutputRecords                int    `json:"outputRecords"`
	ShuffleRemoteBlocksFetched   int    `json:"shuffleRemoteBlocksFetched"`
	ShuffleLocalBlocksFetched    int    `json:"shuffleLocalBlocksFetched"`
	ShuffleFetchWaitTime         int    `json:"shuffleFetchWaitTime"`
	ShuffleRemoteBytesRead       int    `json:"shuffleRemoteBytesRead"`
	ShuffleRemoteBytesReadToDisk int    `json:"shuffleRemoteBytesReadToDisk"`
	ShuffleLocalBytesRead        int    `json:"shuffleLocalBytesRead"`
	ShuffleReadBytes             int    `json:"shuffleReadBytes"`
	ShuffleReadRecords           int    `json:"shuffleReadRecords"`
	ShuffleWriteBytes            int    `json:"shuffleWriteBytes"`
	ShuffleWriteTime             int    `json:"shuffleWriteTime"`
	ShuffleWriteRecords          int    `json:"shuffleWriteRecords"`
	Name                         string `json:"name"`
	Description                  string `json:"description"`
	Details                      string `json:"details"`
	SchedulingPool               string `json:"schedulingPool"`
	ResourceProfileId            int    `json:"resourceProfileId"`
}
