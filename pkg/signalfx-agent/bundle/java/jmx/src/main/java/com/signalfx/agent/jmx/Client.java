package com.signalfx.agent.jmx;

import org.apache.commons.lang3.StringUtils;

import java.io.IOException;
import java.net.MalformedURLException;
import java.security.Security;
import java.util.HashMap;
import java.util.Map;
import java.util.Set;
import java.util.logging.Level;
import java.util.logging.Logger;

import javax.management.MBeanServerConnection;
import javax.management.ObjectName;
import javax.management.remote.JMXConnector;
import javax.management.remote.JMXConnectorFactory;
import javax.management.remote.JMXServiceURL;

public class Client {
    private static Logger logger = Logger.getLogger(Client.class.getName());

    private final JMXServiceURL url;
    private final String username;
    private final String password;
    private final String realm;
    private final String jmxRemoteProfiles;
    private JMXConnector jmxConn;

    Client(JMXMonitor.JMXConfig conf) throws MalformedURLException {
        this.url = new JMXServiceURL(conf.serviceURL);
        this.username = conf.username;
        this.password = conf.password;
        this.realm = conf.realm;
        this.jmxRemoteProfiles = conf.jmxRemoteProfiles;
    }

    private MBeanServerConnection ensureConnected() {
        if (jmxConn != null) {
            try {
                return jmxConn.getMBeanServerConnection();
            } catch (IOException e) {
                // Go on and reestablish the connection below if this is reached.
            }
        }
        try {
            Map<String,Object> env = new HashMap<>();
            if (StringUtils.isNotBlank(username)) {
                env.put(JMXConnector.CREDENTIALS, new String[]{this.username, this.password});
            }
            env.put("jmx.remote.profiles", this.jmxRemoteProfiles);
            env.put("jmx.remote.sasl.callback.handler", new ClientCallbackHandler(this.username, this.password, this.realm));
            Security.addProvider(new com.sun.security.sasl.Provider());
            jmxConn = JMXConnectorFactory.connect(url, env);
            return jmxConn.getMBeanServerConnection();
        } catch (IOException e) {
            logger.log(Level.WARNING, "Could not connect to remote JMX server: ", e);
            return null;
        }
    }

    public MBeanServerConnection getConnection() {
        return ensureConnected();
    }

    public Set<ObjectName> query(ObjectName objectName) {
        MBeanServerConnection mbsc = ensureConnected();
        if (mbsc == null) {
            return null;
        }

        try {
             return mbsc.queryNames(objectName, null);
        } catch (IOException e) {
            jmxConn = null;
            return null;
        }
    }
}
