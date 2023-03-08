package com.signalfx.agent;

import java.util.HashMap;
import java.util.Map;
import java.util.logging.Level;
import java.util.logging.LogManager;
import java.util.logging.Logger;

import com.signalfx.agent.logging.Handler;

public class SignalFxMonitorRunner<T extends MonitorConfig> {
    private static Logger logger = Logger.getLogger(SignalFxMonitorRunner.class.getName());

    private final SignalFxMonitor monitor;
    private final Class configCls;

    public SignalFxMonitorRunner(SignalFxMonitor<T> monitor, Class configCls) {
        this.monitor = monitor;
        this.configCls = configCls;
    }

    public void run() {
        PipeIO agentIO = new PipeIO(System.in, System.out);

        AgentOutput output = new AgentOutput(agentIO);

        setUpLogging(output);

        Message<? extends MonitorConfig> firstMsg = agentIO.receive(this.configCls);

        if (firstMsg.type != MessageType.CONFIGURE) {
            throw new RuntimeException(
                    "Got unexpected first message from agent: " + firstMsg.toString());
        }

        try {
            monitor.configure(firstMsg.payload, output);
        } catch (ConfigureError e) {
            logger.log(Level.SEVERE, "Failed to configure monitor", e);
            Map<String, Object> result = new HashMap();
            result.put("error", e.toString());
            agentIO.send(MessageType.CONFIGURE_RESULT, result);
            return;
        }
        agentIO.send(MessageType.CONFIGURE_RESULT, new HashMap<>());

        output.setReady();

        Message shutdownMsg = agentIO.receive();
        if (shutdownMsg.type != MessageType.SHUTDOWN) {
            throw new RuntimeException(
                    String.format("Expected shutdown message, got message of type %d",
                            shutdownMsg.type));
        }
        monitor.shutdown();
        return;
    }

    public void runDebug() {

    }

    private void setUpLogging(AgentOutput output) {
        Logger root = LogManager.getLogManager().getLogger("");
        java.util.logging.Handler handler = new Handler(output);
        for (java.util.logging.Handler h : root.getHandlers()) {
            root.removeHandler(h);
        }
        root.addHandler(handler);

        Thread.setDefaultUncaughtExceptionHandler((Thread t, Throwable e) ->
                logger.log(Level.SEVERE, "Caught exception in thread " + t.getName(), e));
    }
}
