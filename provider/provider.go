package provider

const (
	// StatusInstanceNotAvailable ...
	StatusInstanceNotAvailable = iota
	// StatusInstanceStarting ...
	StatusInstanceStarting
	// StatusInstanceNotRun ...
	StatusInstanceNotRun
	// StatusInstanceStopping ...
	StatusInstanceStopping
	// StatusInstanceRunning ...
	StatusInstanceRunning
	// StatusInstanceError ...
	StatusInstanceError
)

// Provider ..
type Provider interface {
	String() string
	Hash() string
	GetStatus() (int, error)
	GetIP() (string, error)
	Start() error
	Stop() error
}
