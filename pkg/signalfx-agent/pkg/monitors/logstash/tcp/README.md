# Development Notes

To run Logstash in Docker, but using the host's network:

`$ docker run --net=host -d -v $(pwd)/tmp/pipeline:/usr/share/logstash/pipeline/ --name logstash -e XPACK_MONITORING_ENABLED=false -v /:/hostfs:ro --user root -i docker.elastic.co/logstash/logstash:7.3.0`

This will also mount in a directory for you to put pipeline configuration.  An
example config that has both meter and timer metrics is:

```
input {
  stdin {}
}

filter {
  if [message] =~ /Took .* seconds/ {
	dissect {
	  mapping => {
		"message" => "Took %{duration} seconds"
	  }
	  convert_datatype => {
		"duration" => "float"
	  }
	}
    if "_dissectfailure" not in [tags] {
	  metrics {
	    timer => { "process_time" => "%{duration}" }
	    flush_interval => 10
		# This makes the timing stats pertain to only the previous 5 minutes
		# instead of since Logstash last started.
	    clear_interval => 300
	    add_field => {"app" => "myapp"}
	    add_tag => "metric"
	    add_tag => "other"
	  }
	}
  }
  if [message] == "Logged in" {
    metrics {
      # This determines how often metric events will be sent to the agent, and
	  # thus how often datapoints will be emitted.
      flush_interval => 10
      # The name of the meter will be used to construct the name of the metric
	  # in SignalFx.  For this example, a datapoint called `logins.count` would
	  # be generated.
      meter => "logins"
      add_tag => "metric"
    }
  }
}

output {
  stdout { codec => rubydebug }

  # Only expose metric events to the agent
  if "metric" in [tags] {
    tcp {
      port => 8900
      mode => "server"
      host => "0.0.0.0"
    }
  }
}
```
