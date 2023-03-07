package com.signalfx.agent;

import java.io.DataInputStream;
import java.io.DataOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.Map;

import com.fasterxml.jackson.databind.ObjectMapper;

public class PipeIO {
    private final DataInputStream input;
    private final DataOutputStream output;

    public PipeIO(InputStream input, OutputStream output) {
        this.input = new DataInputStream(input);
        this.output = new DataOutputStream(output);

    }

    public Message receive() {
        return receive(null);
    }

    public <T> Message<T> receive(Class<T> cls) {
        try {
            int type = this.input.readInt();
            int size = this.input.readInt();

            // In case size gets misinterpreted due to stream corruption make sure we don't blow up memory
            // by creating a crazy big array.  No message coming in should ever be more than 1MB.
            assert size < 1024 * 1024 : String.format("Size was too big: %d", size);


            T payload = null;
            if (size > 0) {
                byte[] readBuffer = new byte[size];
                this.input.readFully(readBuffer);
                payload = deserialize(readBuffer, cls);
            }

            return new Message<T>(type, size, payload);
        } catch (IOException e) {
            throw new RuntimeException("Failed to receive from agent", e);
        }
    }

    private <T> T deserialize(byte[] buf, Class<T> cls) throws IOException {
        ObjectMapper mapper = new ObjectMapper();
        return mapper.readValue(buf, cls);
    }

    public void send(int type, byte[] payload) {
        try {
            this.output.writeInt(type);
            this.output.writeInt(payload.length);
            this.output.write(payload);
        } catch (IOException e) {
            throw new RuntimeException("Failed to send to agent", e);
        }
    }

    public void send(int type, Map<String, Object> payloadObj) {
        ObjectMapper mapper = new ObjectMapper();

        byte[] payload;
        try {
            payload = mapper.writeValueAsBytes(payloadObj);
        } catch (IOException e) {
            throw new RuntimeException("Failed to serialize to JSON", e);
        }

        send(type, payload);
    }
}
