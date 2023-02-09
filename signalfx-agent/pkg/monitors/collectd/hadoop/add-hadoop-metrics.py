#!/usr/bin/env python3
#
# Super hacky script to add hadoop metadata.
#
import sys
from ruamel import yaml

# Below taken from https://github.com/signalfx/collectd-hadoop/blob/master/metrics.py
############

GAUGE = "gauge"
COUNTER = "counter"

METRICS = {}

HADOOP_CLUSTER_METRICS = {
    "activeNodes": (GAUGE, "hadoop.cluster.metrics.active_nodes"),
    "allocatedMB": (GAUGE, "hadoop.cluster.metrics.allocated_mb"),
    "allocatedVirtualCores": (GAUGE, "hadoop.cluster.metrics.allocated_virtual_cores"),
    "appsCompleted": (GAUGE, "hadoop.cluster.metrics.apps_completed"),
    "appsFailed": (GAUGE, "hadoop.cluster.metrics.apps_failed"),
    "appsKilled": (GAUGE, "hadoop.cluster.metrics.apps_killed"),
    "appsPending": (GAUGE, "hadoop.cluster.metrics.apps_pending"),
    "appsRunning": (GAUGE, "hadoop.cluster.metrics.apps_running"),
    "appsSubmitted": (GAUGE, "hadoop.cluster.metrics.apps_submitted"),
    "availableMB": (GAUGE, "hadoop.cluster.metrics.available_mb"),
    "availableVirtualCores": (GAUGE, "hadoop.cluster.metrics.available_virtual_cores"),
    "containersAllocated": (GAUGE, "hadoop.cluster.metrics.containers_allocated"),
    "containersPending": (GAUGE, "hadoop.cluster.metrics.containers_pending"),
    "containersReserved": (GAUGE, "hadoop.cluster.metrics.containers_reserved"),
    "decommissionedNodes": (GAUGE, "hadoop.cluster.metrics.decommissioned_nodes"),
    "lostNodes": (GAUGE, "hadoop.cluster.metrics.lost_nodes"),
    "rebootedNodes": (GAUGE, "hadoop.cluster.metrics.rebooted_nodes"),
    "reservedMB": (GAUGE, "hadoop.cluster.metrics.reserved_mb"),
    "reservedVirtualCores": (GAUGE, "hadoop.cluster.metrics.reserved_virtual_cores"),
    "totalMB": (COUNTER, "hadoop.cluster.metrics.total_mb"),
    "totalNodes": (COUNTER, "hadoop.cluster.metrics.total_nodes"),
    "totalVirtualCores": (COUNTER, "hadoop.cluster.metrics.total_virtual_cores"),
    "unhealthyNodes": (GAUGE, "hadoop.cluster.metrics.unhealthy_nodes"),
}

HADOOP_LEAF_QUEUE = {
    "absoluteCapacity": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.absoluteCapacity"),
    "absoluteMaxCapacity": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.absoluteMaxCapacity"),
    "absoluteUsedCapacity": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.absoluteUsedCapacity"),
    "capacity": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.capacity"),
    "maxActiveApplications": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.maxActiveApplications"),
    "maxActiveApplicationsPerUser": (
        GAUGE,
        "hadoop.resource.manager.scheduler.leaf.queue.maxActiveApplicationsPerUser",
    ),
    "maxApplications": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.maxApplications"),
    "maxApplicationsPerUser": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.maxApplicationsPerUser"),
    "maxCapacity": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.maxCapacity"),
    "numActiveApplications": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.numActiveApplications"),
    "numApplications": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.numApplications"),
    "numContainers": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.numContainers"),
    "numPendingApplications": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.numPendingApplications"),
    "usedCapacity": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.usedCapacity"),
    "userLimit": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.userLimit"),
    "userLimitFactor": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.userLimitFactor"),
    "allocatedContainers": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.allocatedContainers"),
    "reservedContainers": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.reservedContainers"),
    "pendingContainers": (GAUGE, "hadoop.resource.manager.scheduler.leaf.queue.pendingContainers"),
}

HADOOP_ROOT_QUEUE = {
    "capacity": (GAUGE, "hadoop.resource.manager.scheduler.root.queue.capacity"),
    "usedCapacity": (GAUGE, "hadoop.resource.manager.scheduler.root.queue.usedCapacity"),
    "maxCapacity": (GAUGE, "hadoop.resource.manager.scheduler.root.queue.maxCapacity"),
}

HADOOP_QUEUE_USERS = {
    "numActiveApplications": (GAUGE, "hadoop.resource.manager.scheduler.queue.users.numActiveApplications"),
    "numPendingApplications": (GAUGE, "hadoop.resource.manager.scheduler.queue.users.numPendingApplications"),
}

HADOOP_RESOURCE_OBJECT = {
    "memory": (GAUGE, "hadoop.resource.manager.scheduler.queue.resource.memory"),
    "vCores": (GAUGE, "hadoop.resource.manager.scheduler.queue.resource.vCores"),
}

HADOOP_FIFO_SCHEDULER = {
    "capacity": (GAUGE, "hadoop.resource.manager.scheduler.fifo.capacity"),
    "usedCapacity": (GAUGE, "hadoop.resource.manager.scheduler.fifo.usedCapacity"),
    "minQueueMemoryCapacity": (GAUGE, "hadoop.resource.manager.scheduler.fifo.minQueueMemoryCapacity"),
    "maxQueueMemoryCapacity": (GAUGE, "hadoop.resource.manager.scheduler.fifo.maxQueueMemoryCapacity"),
    "numNodes": (GAUGE, "hadoop.resource.manager.scheduler.fifo.numNodes"),
    "usedNodeCapacity": (GAUGE, "hadoop.resource.manager.scheduler.fifo.usedNodeCapacity"),
    "availNodeCapacity": (GAUGE, "hadoop.resource.manager.scheduler.fifo.availNodeCapacity"),
    "totalNodeCapacity": (GAUGE, "hadoop.resource.manager.scheduler.fifo.totalNodeCapacity"),
    "numContainers": (GAUGE, "hadoop.resource.manager.scheduler.fifo.numContainers"),
}

HADOOP_APPLICATIONS = {
    "progress": (GAUGE, "hadoop.resource.manager.apps.progress"),
    "priority": (GAUGE, "hadoop.resource.manager.apps.priority"),
    "allocatedMB": (GAUGE, "hadoop.resource.manager.apps.allocatedMB"),
    "allocatedVCores": (GAUGE, "hadoop.resource.manager.apps.allocatedVCores"),
    "runningContainers": (GAUGE, "hadoop.resource.manager.apps.runningContainers"),
    "memorySeconds": (GAUGE, "hadoop.resource.manager.apps.memorySeconds"),
    "vcoreSeconds": (GAUGE, "hadoop.resource.manager.apps.vcoreSeconds"),
    "queueUsagePercentage": (GAUGE, "hadoop.resource.manager.apps.queueUsagePercentage"),
    "clusterUsagePercentage": (GAUGE, "hadoop.resource.manager.apps.clusterUsagePercentage"),
    "preemptedResourceMB": (GAUGE, "hadoop.resource.manager.apps.preemptedResourceMB"),
    "preemptedResourceVCores": (GAUGE, "hadoop.resource.manager.apps.preemptedResourceVCores"),
    "numNonAMContainerPreempted": (GAUGE, "hadoop.resource.manager.apps.numNonAMContainerPreempted"),
    "numAMContainerPreempted": (GAUGE, "hadoop.resource.manager.apps.numAMContainerPreempted"),
}

HADOOP_NODE_METRICS = {
    "numContainers": (GAUGE, "hadoop.resource.manager.nodes.numContainers"),
    "usedMemoryMB": (GAUGE, "hadoop.resource.manager.nodes.usedMemoryMB"),
    "availMemoryMB": (GAUGE, "hadoop.resource.manager.nodes.availMemoryMB"),
    "usedVirtualCores": (GAUGE, "hadoop.resource.manager.nodes.usedVirtualCores"),
    "availableVirtualCores": (GAUGE, "hadoop.resource.manager.nodes.availableVirtualCores"),
}

HADOOP_NODE_RESOURCE_UTIL = {
    "nodePhysicalMemoryMB": (GAUGE, "hadoop.resource.manager.node.nodePhysicalMemoryMB"),
    "nodeVirtualMemoryMB": (GAUGE, "hadoop.resource.manager.node.nodeVirtualMemoryMB"),
    "nodeCPUUsage": (GAUGE, "hadoop.resource.manager.node.nodeCPUUsage"),
}


MAPREDUCE_JOB_METRICS = {
    "elapsedTime": (GAUGE, "hadoop.mapreduce.job.elapsedTime"),
    "mapsTotal": (GAUGE, "hadoop.mapreduce.job.mapsTotal"),
    "mapsCompleted": (GAUGE, "hadoop.mapreduce.job.mapsCompleted"),
    "reducesTotal": (GAUGE, "hadoop.mapreduce.job.reducesTotal"),
    "reducesCompleted": (GAUGE, "hadoop.mapreduce.job.reducesCompleted"),
    "mapsPending": (GAUGE, "hadoop.mapreduce.job.mapsPending"),
    "mapsRunning": (GAUGE, "hadoop.mapreduce.job.mapsRunning"),
    "reducesPending": (GAUGE, "hadoop.mapreduce.job.reducesPending"),
    "reducesCompleted": (GAUGE, "hadoop.mapreduce.job.reducesCompleted"),
    "newReduceAttempts": (GAUGE, "hadoop.mapreduce.job.newReduceAttempts"),
    "runningReduceAttempts": (GAUGE, "hadoop.mapreduce.job.runningReduceAttempts"),
    "failedReduceAttempts": (GAUGE, "hadoop.mapreduce.job.failedReduceAttempts"),
    "killedReduceAttempts": (GAUGE, "hadoop.mapreduce.job.killedReduceAttempts"),
    "successfulReduceAttempts": (GAUGE, "hadoop.mapreduce.job.successfulReduceAttempts"),
    "newMapAttempts": (GAUGE, "hadoop.mapreduce.job.newMapAttempts"),
    "runningMapAttempts": (GAUGE, "hadoop.mapreduce.job.runningMapAttempts"),
    "failedMapAttempts": (GAUGE, "hadoop.mapreduce.job.failedMapAttempts"),
    "killedMapAttempts": (GAUGE, "hadoop.mapreduce.job.killedMapAttempts"),
    "successfulMapAttempts": (GAUGE, "hadoop.mapreduce.job.successfulMapAttempts"),
}


###################

METRICS["cluster"] = HADOOP_CLUSTER_METRICS
METRICS["leaf-queue"] = HADOOP_LEAF_QUEUE
METRICS["root-queue"] = HADOOP_ROOT_QUEUE
METRICS["queue-users"] = HADOOP_QUEUE_USERS
METRICS["resource-objects"] = HADOOP_RESOURCE_OBJECT
METRICS["fifo-scheduler"] = HADOOP_FIFO_SCHEDULER
METRICS["applications"] = HADOOP_APPLICATIONS
METRICS["nodes"] = HADOOP_NODE_METRICS
METRICS["node-resources"] = HADOOP_NODE_RESOURCE_UTIL
METRICS["mapreduce-jobs"] = MAPREDUCE_JOB_METRICS


def map_metric_type(typ):
    typ = typ.tolower()
    if typ in ("cumulative", "gauge"):
        return typ
    return dict(CUMULATIVE_COUNTER="cumulative", GAUGE="gauge")[typ]


metaFile = sys.argv[1]

with open(metaFile) as f:
    meta = yaml.round_trip_load(f)

assert len(meta["monitors"]) == 1, "only supports 1 monitor"
monitor = meta["monitors"][0]
metrics = monitor["metrics"]

for group, hadoop_metrics in METRICS.items():
    monitor.setdefault("groups", {}).setdefault(group, {"description": None})

    for info in hadoop_metrics.values():
        typ, metric = info

        metrics.setdefault(metric, {"description": None, "type": typ, "included": False, "group": group})
        assert metrics[metric]["type"] == typ, f"{metric} type doesn't match observed metric type {typ}"

with open(metaFile, "wt") as f:
    yaml.round_trip_dump(meta, f)
