package state

// InMemory stores the state in memory. It is hence volatile and only persistent as long
// as the program is running.
type InMemory struct {
	stored *MonitoredIPs
}

// NewInMemory creates a new in-memory state
func NewInMemory() *InMemory {
	return &InMemory{
		stored: &MonitoredIPs{},
	}
}

// Set sets the state to ips
func (m *InMemory) Set(ips MonitoredIPs) error {
	m.stored = &ips
	return nil
}

// Get returns the currently stored state
func (m *InMemory) Get() (MonitoredIPs, error) {
	return *m.stored, nil
}

// Reset returns the state to its default value
func (m *InMemory) Reset() error {
	m.stored = &MonitoredIPs{}
	return nil
}
