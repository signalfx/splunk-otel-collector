package services

import "fmt"

// OrchestrationType represents the type of orchestration the service is
// deployed under.
type OrchestrationType int

const (
	// KUBERNETES orchestrator
	KUBERNETES OrchestrationType = 1 + iota
	// MESOS orchestrator
	MESOS
	// SWARM orchestrator
	SWARM
	// DOCKER orchestrator
	DOCKER
	// ECS orchestrator
	ECS
	// NONE orchestrator
	NONE
)

// PortPreference describes whether the public or private port should be preferred
// when connecting to the service
type PortPreference int

const (
	// PRIVATE means that the internal port (e.g. what the service is
	// configured to listen on directly) should be used when connecting
	PRIVATE PortPreference = 1 + iota
	// PUBLIC means that the port that is exposed through some network mapping
	// should be used (e.g. via Docker's -p flag)
	PUBLIC
)

// Orchestration contains information about the orchestrator that the service
// is deployed on (see OrchestrationType)
type Orchestration struct {
	ID       string            `yaml:"-"`
	Type     OrchestrationType `yaml:"orchestrator"`
	PortPref PortPreference    `yaml:"-"`
}

// NewOrchestration constructor
func NewOrchestration(id string, orchType OrchestrationType, portPref PortPreference) *Orchestration {
	return &Orchestration{id, orchType, portPref}
}

func (o *Orchestration) String() string {
	return fmt.Sprintf("%#v", o)
}
