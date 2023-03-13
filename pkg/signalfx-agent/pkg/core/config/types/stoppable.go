package types

// Stoppable is something that can be stopped
type Stoppable interface {
	Stop() error
}
