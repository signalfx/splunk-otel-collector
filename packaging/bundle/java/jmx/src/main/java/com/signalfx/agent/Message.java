package com.signalfx.agent;

public class Message<T> {
    public final int type;
    public final int size;
    public final T payload;

    public Message(int type, int size,
                   T payload) {
        this.type = type;
        this.size = size;
        this.payload = payload;
    }

    public String toString() {
        return String.format("Message<type=%d; size=%d; payload='%s'>", type, size, payload);
    }
}
