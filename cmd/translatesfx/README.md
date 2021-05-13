# SignalFx Smart Agent Configuration Translation Tool (Experimental)

This package provides a command-line tool to translate a SignalFx Smart Agent
configuration file into an OpenTelemetry Collector configuration.

## Usage

The `translatesfx` command requires one argument, the signalfx configuration
file, and accepts a second argument, the working directory used by any #from
directives. The `translatesfx` command uses the working directory to resolve any
relative paths to files. If you omit the working directory argument, 
`translatesfx` expands relative files paths using the current working
directory.

```
> translatesfx <sfx-file> [<file expansion working directory>]
```

## Examples

###### Using the current working directory as the working directory:

```
> translatesfx path/to/sfx/config.yaml
```

###### Using a custom working directory:

```
> translatesfx path/to/sfx/config.yaml path/to/sfx
```
