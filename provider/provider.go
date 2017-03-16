package provider

// StatusInstance ...
type StatusInstance int

const (
	// StatusInstanceNotAvailable ...
	StatusInstanceNotAvailable StatusInstance = iota
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

func (status StatusInstance) String() string {
	switch status {
	case StatusInstanceNotAvailable:
		return "not available"
	case StatusInstanceStarting:
		return "starting"
	case StatusInstanceNotRun:
		return "not run"
	case StatusInstanceStopping:
		return "stopping"
	case StatusInstanceRunning:
		return "running"
	case StatusInstanceError:
		return "error"
	}

	return "unknown"
}

// Provider ..
type Provider interface {
	String() string
	Hash() string
	Status() (StatusInstance, error)
	IP() (string, error)
	Start() error
	Stop() error
}
