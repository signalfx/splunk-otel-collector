// Package monitors is the core logic for monitors.  Monitors are what collect
// metrics from the environment.  They have a simple interface that all must
// implement: the Configure method, which takes one argument of the same type
// that you pass as the configTemplate to the Register function.  Optionally,
// monitors may implement the niladic Shutdown method to do cleanup.
//
// Monitors will never be reused after the Shutdown method is called.
//
// The new monitor instance will be created for each service monitored.  The
// Configure method will be called at most one time after the monitor is
// instantiated.  Monitors should start any long running routines in the
// Configure method and start monitoring immediately.
//
// Monitors are responsible for sending datapoints on the configured interval
// (the IntervalSeconds field of the config struct).  There is no "read"
// callback method that will be called, so it is up to the monitors to maintain
// a timer however they wish to comply with the reporting interval.
//
// If a monitor wants to create SignalFx golib datapoints/events and have them
// sent by the agent.  The monitor type should define a "DPs" and/or "Events"
// field of the type "chan<- datapoints.Datapoint" and "chan<- events.Event".
// The monitor manager will automatically inject those fields before Configure
// is called.  They could be swapped out at any time, so monitors should not
// cache those fields in other variables.
//
// Monitors can also specify a field on their main monitor type (NOT their
// config) called "AgentMeta" of the type AgentMeta that provides monitors with
// access to agent global information, such as the hostname.  Note that this
// information returned by the AgentMeta methods could change at any time, so
// monitors should call those methods each time it needs that information
// instead of caching it.
package monitors
