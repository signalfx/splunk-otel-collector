package com.signalfx.agent;

public interface SignalFxMonitor<T extends MonitorConfig> {
     void configure(T conf, AgentOutput output);
     void shutdown();
}
