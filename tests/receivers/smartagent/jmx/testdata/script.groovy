ss = util.queryJMX("org.apache.cassandra.db:type=StorageService").first()

// Copied and modified from https://github.com/apache/cassandra
def parseFileSize(String value) {
  if (!value.matches("\\d+(\\.\\d+)? (GiB|KiB|MiB|TiB|bytes)")) {
    throw new IllegalArgumentException(
      String.format("value %s is not a valid human-readable file size", value));
  }
  if (value.endsWith(" TiB")) {
    return Math.round(Double.valueOf(value.replace(" TiB", "")) * 1e12);
  }
  else if (value.endsWith(" GiB")) {
    return Math.round(Double.valueOf(value.replace(" GiB", "")) * 1e9);
  }
  else if (value.endsWith(" KiB")) {
    return Math.round(Double.valueOf(value.replace(" KiB", "")) * 1e3);
  }
  else if (value.endsWith(" MiB")) {
    return Math.round(Double.valueOf(value.replace(" MiB", "")) * 1e6);
  }
  else if (value.endsWith(" bytes")) {
    return Math.round(Double.valueOf(value.replace(" bytes", "")));
  }
  else {
    throw new IllegalStateException(String.format("FileUtils.parseFileSize() reached an illegal state parsing %s", value));
  }
}

localEndpoint = ss.HostIdToEndpoint.get(ss.LocalHostId)
log.info(localEndpoint)
dims = [host_id: ss.LocalHostId, cluster_name: ss.ClusterName]

// Equivalent of "Up/Down" in the `nodetool status` output.
// 1 = Live; 0 = Dead; -1 = Unknown
output.sendDatapoint(util.makeGauge(
  "cassandra.status",
  ss.LiveNodes.contains(localEndpoint) ? 1 : (ss.DeadNodes.contains(localEndpoint) ? 0 : -1),
  dims))

output.sendDatapoint(util.makeGauge(
  "cassandra.state",
  ss.JoiningNodes.contains(localEndpoint) ? 3 : (ss.LeavingNodes.contains(localEndpoint) ? 2 : 1),
  dims))

output.sendDatapoint(util.makeGauge(
  "cassandra.load",
  parseFileSize(ss.LoadString),
  dims))

log.info(ss.Ownership.toString())
output.sendDatapoint(util.makeGauge(
  "cassandra.ownership",
  ss.Ownership.get(InetAddress.getByName(localEndpoint)),
  dims))

log.info("groovy init script complete")