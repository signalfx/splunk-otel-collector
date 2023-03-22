package meta

// AgentMeta provides monitors access to global agent metadata.  Putting this
// into a single interface allows easy expansion of metadata without breaking
// backwards-compatibility and without exposing global variables that monitors
// access.
// TODO: get rid of this since it's hacky
type AgentMeta struct {
	InternalStatusHost string
	InternalStatusPort uint16
}
