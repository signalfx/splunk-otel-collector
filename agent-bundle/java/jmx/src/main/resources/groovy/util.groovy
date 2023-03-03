import com.signalfx.agent.MonitorUtil
import com.signalfx.agent.jmx.Client
import com.signalfx.metrics.protobuf.SignalFxProtocolBuffers

import javax.management.MBeanServerConnection
import javax.management.ObjectName

class SignalFx {
    private final Client jmxClient;

    SignalFx(Client jmxClient) {
        this.jmxClient = jmxClient;
    }

    List<GroovyMBean> queryJMX(String objNameStr) {
        ObjectName objName = new ObjectName(objNameStr)
        Set<ObjectName> names = jmxClient.query(objName)
        MBeanServerConnection server = jmxClient.getConnection()

        return names.collect{ new GroovyMBean(server, it) }
    }

    SignalFxProtocolBuffers.DataPoint makeGauge(String name, double val, Map<String, String> dims) {
        return MonitorUtil.makeGauge(name, val, dims.collect {
            MonitorUtil.newDim(it.key, it.value)
        })
    }

    SignalFxProtocolBuffers.DataPoint makeCumulative(String name, double val, Map<String, String> dims) {
        return MonitorUtil.makeCumulative(name, val, dims.collect {
            MonitorUtil.newDim(it.key, it.value)
        })
    }
}