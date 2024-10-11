package com.signalfx.agent;

public class ConfigureError extends RuntimeException {
    public ConfigureError(String message, Throwable cause) {
        super(message, cause);
    }
}
