package com.signalfx.agent.jmx;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.lang.reflect.InvocationTargetException;
import java.util.logging.Level;
import java.util.logging.Logger;
import java.util.stream.Collectors;

import org.codehaus.groovy.control.CompilationFailedException;

import com.signalfx.agent.AgentOutput;
import com.signalfx.agent.ConfigureError;


import groovy.lang.Binding;
import groovy.lang.GroovyShell;
import groovy.lang.Script;

public class GroovyRunner {
    private static Logger logger = Logger.getLogger(GroovyRunner.class.getName());

    private final GroovyShell gshell = new GroovyShell();

    private final Binding binding;
    private final Script script;
    private final AgentOutput output;

    GroovyRunner(String groovyScript, AgentOutput output, Client jmxClient) {
        try {
            this.script = gshell.parse(groovyScript);
        } catch (CompilationFailedException e) {
            logger.log(Level.SEVERE, "Failed to compile groovy script", e);
            throw new ConfigureError("Failed to compile groovy script", e);
        }
        this.output = output;

        binding = new Binding();

        binding.setVariable("output", output);
        binding.setVariable("log", logger);

        script.setBinding(binding);

        // The util script provides some helpers to make it easy to query JMX and make datapoints in Groovy.
        try {
            String utilScriptText = getResourceFileAsString("groovy/util.groovy");
            Object utilInstance = gshell.getClassLoader().parseClass(utilScriptText)
                    .getDeclaredConstructor(Client.class).newInstance(jmxClient);
            binding.setVariable("util", utilInstance);
        } catch (IOException | InstantiationException | IllegalAccessException | NoSuchMethodException | InvocationTargetException e) {
            throw new RuntimeException("Could not load util groovy from jar", e);
        }
    }

    static String getResourceFileAsString(String fileName) throws IOException {
        ClassLoader classLoader = ClassLoader.getSystemClassLoader();
        try (InputStream is = classLoader.getResourceAsStream(fileName)) {
            if (is == null)
                return null;
            try (InputStreamReader isr = new InputStreamReader(is);
                    BufferedReader reader = new BufferedReader(isr)) {
                return reader.lines().collect(Collectors.joining(System.lineSeparator()));
            }
        }
    }

    public void run() {
        script.run();
    }
}
