package ecs

import (
	"fmt"
	"strings"
)

// TaskMetadata contains a set of properties from ECS Task Metadata
type TaskMetadata struct {
	ClusterName string      `json:"Cluster"`
	TaskARN     string      `json:"TaskARN"`
	Family      string      `json:"Family"`
	Revision    string      `json:"Revision"`
	KnownStatus string      `json:"KnownStatus"`
	Containers  []Container `json:"Containers"`
}

// GetDimensions returns a set of dimensions based on the task metadata
func (task *TaskMetadata) GetDimensions() map[string]string {
	dims := make(map[string]string)
	clusterNameLen := len(task.ClusterName)
	if idx := strings.Index(task.ClusterName, "/"); idx >= 0 && idx+1 < clusterNameLen {
		dims["ClusterName"] = task.ClusterName[idx+1 : clusterNameLen]
	} else {
		dims["ClusterName"] = task.ClusterName
	}
	dims["ecs_cluster_arn"] = task.ClusterName
	dims["ecs_task_arn"] = task.TaskARN
	arnBaseIdx := strings.Index(task.TaskARN, ":task/")
	arnBase := task.TaskARN[0:arnBaseIdx]
	dims["ecs_task_definition_arn"] = fmt.Sprintf("%s:task-definition/%s:%s", arnBase, task.Family, task.Revision)
	dims["ecs_task_group"] = task.Family
	dims["ecs_task_version"] = task.Revision

	return dims
}

// Container struct represents container structure that is a part of ECS Task Metadata
type Container struct {
	DockerID    string            `json:"DockerId"`
	Name        string            `json:"DockerName"`
	Image       string            `json:"Image"`
	KnownStatus string            `json:"KnownStatus"`
	Type        string            `json:"Type"`
	Labels      map[string]string `json:"Labels"`
	Limits      struct {
		CPU int64 `json:"CPU"`
	}
	Networks []struct {
		NetworkMode string   `json:"NetworkMode"`
		IPAddresses []string `json:"IPv4Addresses"`
	}
}
