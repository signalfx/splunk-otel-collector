package com.signalfx.agent.jmx;

import java.net.MalformedURLException;
import java.util.Timer;
import java.util.logging.Level;
import java.util.logging.Logger;

import com.signalfx.agent.AgentOutput;
import com.signalfx.agent.ConfigureError;
import com.signalfx.agent.MonitorConfig;
import com.signalfx.agent.SignalFxMonitor;
import com.signalfx.agent.SignalFxMonitorRunner;
import com.signalfx.agent.MonitorUtil;
import org.apache.commons.lang3.StringUtils;

public class JMXMonitor implements SignalFxMonitor<JMXMonitor.JMXConfig> {
    private static Logger logger = Logger.getLogger(JMXMonitor.class.getName());

    private final Timer timer = new Timer();

    public static class JMXConfig extends MonitorConfig {
        public String serviceURL;
        public String groovyScript;
        public String username;
        public String password;
        public String keyStorePath;
        public String keyStorePassword;
        public String keyStoreType;
        public String trustStorePath;
        public String trustStorePassword;
        public String jmxRemoteProfiles;
        public String realm;
    }

    public void configure(JMXConfig conf, AgentOutput output) {
        setSystemProperties(conf);
        Client client;
        try {
            client = new Client(conf);
        } catch(MalformedURLException e) {
            throw new ConfigureError("Malformed serviceUrl: ", e);
        }

        GroovyRunner runner = new GroovyRunner(conf.groovyScript, output, client);

        timer.scheduleAtFixedRate(MonitorUtil.wrapTimerTask(() -> {
            try {
                runner.run();
            } catch (Throwable e) {
                logger.log(Level.SEVERE, "Error gathering JMX metrics", e);
            }

        }), 0, conf.intervalSeconds * 1000);
    }

    private void setSystemProperties(JMXConfig conf) {
        if (StringUtils.isNotBlank(conf.keyStorePath)) {
            System.setProperty("javax.net.ssl.keyStore", conf.keyStorePath);
        }
        if (StringUtils.isNotBlank(conf.keyStorePassword)) {
            System.setProperty("javax.net.ssl.keyStorePassword", conf.keyStorePassword);
        }
        if (StringUtils.isNotBlank(conf.keyStoreType)) {
            System.setProperty("javax.net.ssl.keyStoreType", conf.keyStoreType);
        }
        if (StringUtils.isNotBlank(conf.trustStorePath)) {
            System.setProperty("javax.net.ssl.trustStore", conf.trustStorePath);
        }
        if (StringUtils.isNotBlank(conf.trustStorePassword)) {
            System.setProperty("javax.net.ssl.trustStorePassword", conf.trustStorePassword);
        }
    }

    public void shutdown() {
        timer.cancel();
        return;
    }

    public static void main(String[] args) {
        SignalFxMonitorRunner runner = new SignalFxMonitorRunner(new JMXMonitor(), JMXConfig.class);
        runner.run();
    }
}
