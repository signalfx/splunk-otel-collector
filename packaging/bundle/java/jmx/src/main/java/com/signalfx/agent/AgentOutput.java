package com.signalfx.agent;

import java.util.Collections;
import java.util.List;
import java.util.concurrent.atomic.AtomicBoolean;

import com.signalfx.metrics.protobuf.SignalFxProtocolBuffers;

public class AgentOutput {
    private final PipeIO agentIO;
    private final AtomicBoolean ready;

    public AgentOutput(PipeIO agentIO) {
        this.agentIO = agentIO;
        ready = new AtomicBoolean(false);
    }

    public void setReady() {
        ready.set(true);
    }

    /**
     * Busy-wait for the output to be put into the ready state.  This should happen pretty quickly.
     */
    private void waitForReady() {
        while(!ready.get()) {
            try {
                Thread.sleep(10);
            } catch (InterruptedException e) {
                return;
            }
        }
    }

    public void sendDatapoint(SignalFxProtocolBuffers.DataPoint dp) {
        sendDatapoints(Collections.singletonList(dp));
    }

    public void sendDatapoints(List<SignalFxProtocolBuffers.DataPoint> datapoints) {
        waitForReady();

        byte[] bodyBytes = SignalFxProtocolBuffers.DataPointUploadMessage.newBuilder()
                .addAllDatapoints(datapoints).build().toByteArray();

        agentIO.send(MessageType.DATAPOINT_PROTOBUF_LIST, bodyBytes);
    }

    public void sendLog(String jsonMsg) {
        agentIO.send(MessageType.LOG, jsonMsg.getBytes());

    }
}
