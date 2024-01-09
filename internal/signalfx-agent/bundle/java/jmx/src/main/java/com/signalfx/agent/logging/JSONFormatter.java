package com.signalfx.agent.logging;

import java.util.HashMap;
import java.util.Map;
import java.util.logging.Formatter;
import java.util.logging.LogRecord;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

public class JSONFormatter extends Formatter {
    public String format(LogRecord rec) {
        Map<String, Object> obj = new HashMap();

        String msg = rec.getMessage();
        if (rec.getThrown() != null) {
            msg += "\n" + rec.getThrown().toString();
        }
        obj.put("message", msg);
        obj.put("level", rec.getLevel().toString());
        obj.put("logger", rec.getLoggerName());
        obj.put("source_path", rec.getSourceClassName() + ":" + rec.getSourceMethodName());
        obj.put("created", rec.getMillis());

        ObjectMapper mapper = new ObjectMapper();
        try {
            return new String(mapper.writeValueAsBytes(obj));
        } catch (JsonProcessingException e) {
            System.err.println("Failed to format log: " + e.toString());
            return "";
        }
    }
}
