package com.signalfx.agent.logging;

import java.util.logging.LogRecord;

import com.signalfx.agent.AgentOutput;

public class Handler extends java.util.logging.Handler {
    private final AgentOutput output;

    public Handler(AgentOutput output) {
        this.output = output;
        this.setFormatter(new JSONFormatter());
    }

    public void publish(LogRecord rec) {
        String raw = getFormatter().format(rec);
        output.sendLog(raw);
    }

    public void flush() {
        return;
    }

    public void close() {
        return;
    }
}
