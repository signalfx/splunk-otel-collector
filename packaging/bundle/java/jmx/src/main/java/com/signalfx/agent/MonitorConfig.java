package com.signalfx.agent;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public class MonitorConfig {
    public int intervalSeconds;
}
