package com.signalfx.agent;

import java.util.ArrayList;
import java.util.List;
import java.util.TimerTask;

import com.signalfx.metrics.protobuf.SignalFxProtocolBuffers;

public class MonitorUtil {
    public static TimerTask wrapTimerTask(final Runnable r) {
        return new TimerTask() {
            @Override
            public void run() {
                r.run();
            }
        };
    }

    public static SignalFxProtocolBuffers.Dimension newDim(String key, String value) {
        return SignalFxProtocolBuffers.Dimension.newBuilder().setKey(key).setValue(value).build();
    }

    public static List<SignalFxProtocolBuffers.Dimension> newDims(String... keyVals) {
        assert keyVals.length % 2 == 0 : "args must be a multiple of two";

        List<SignalFxProtocolBuffers.Dimension> dims = new ArrayList(keyVals.length / 2);
        for (int i = 0; i < keyVals.length; i++) {
            if (i % 2 == 1) {
                dims.add(newDim(keyVals[i - 1], keyVals[i]));
            }
        }

        return dims;
    }

    public static SignalFxProtocolBuffers.DataPoint makeDatapoint(String metricName,
                                                                  SignalFxProtocolBuffers.MetricType type,
                                                                  double val,
                                                                  List<SignalFxProtocolBuffers.Dimension> dimensions) {
        SignalFxProtocolBuffers.DataPoint.Builder builder = SignalFxProtocolBuffers.DataPoint
                .newBuilder()
                .setMetric(metricName)
                .setValue(SignalFxProtocolBuffers.Datum.newBuilder().setDoubleValue(val)
                        .build())
                .setMetricType(type);

        if (dimensions != null) {
            builder.addAllDimensions(dimensions);
        }
        return builder.build();
    }

    public static SignalFxProtocolBuffers.DataPoint makeGauge(String metricName, double val,
                                                              List<SignalFxProtocolBuffers.Dimension> dimensions) {
        return makeDatapoint(metricName, SignalFxProtocolBuffers.MetricType.GAUGE, val, dimensions);
    }

    public static SignalFxProtocolBuffers.DataPoint makeCumulative(String metricName, double val,
                                                              List<SignalFxProtocolBuffers.Dimension> dimensions) {
        return makeDatapoint(metricName, SignalFxProtocolBuffers.MetricType.CUMULATIVE_COUNTER, val, dimensions);
    }
}
