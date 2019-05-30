package listener

import (
	"os"
	"reflect"

	yaml "gopkg.in/yaml.v2"
)

// State stores and informs about the current state of the monitored interface's IPs
type State interface {
	Setter
	Getter
	Reset() error
	InSync(MonitoredIPs) bool
}

// Setter allows to set the state to MonitoredIPs
type Setter interface {
	Set(MonitoredIPs) error
}

// Getter allows to retrieve the current state of the MonitoredIPs
type Getter interface {
	Get() (MonitoredIPs, error)
}

// MonitoredIPs stores the IP objects
type MonitoredIPs struct {
	IPv4 string
	IPv6 string
}

// Set writes a the state to disk in YAML representation
func (f *FileState) Set(ips MonitoredIPs) error {
	f.stored = &ips
	return yaml.NewEncoder(f.fd).Encode(f.stored)
}

// Get reads the state from a YAML file from disk
func (f *FileState) Get() (MonitoredIPs, error) {
	err := yaml.NewDecoder(f.fd).Decode(f.stored)
	return *f.stored, err
}

// InSync tests whether the internal state corresponds to the provided one
func (f *FileState) InSync(with MonitoredIPs) bool {
	return reflect.DeepEqual(with, *f.stored)
}

// Reset writes an epty state to disk
func (f *FileState) Reset() error {
	return f.Set(MonitoredIPs{})
}

// FileState supplies methods to handle the state via a file
type FileState struct {
	fd     *os.File
	stored *MonitoredIPs
}

// NewFileState creates a new FileState
func NewFileState(path string) (*FileState, error) {
	var err error

	f := new(FileState)

	f.fd, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// InMemoryState stores the state in memory. It is hence volatile and only persistent as long
// as the program is running.
type InMemoryState struct {
	stored *MonitoredIPs
}

// NewInMemoryState creates a new in-memory state
func NewInMemoryState() *InMemoryState {
	return &InMemoryState{
		stored: &MonitoredIPs{},
	}
}

// Set sets the state to ips
func (m *InMemoryState) Set(ips MonitoredIPs) error {
	m.stored = &ips
	return nil
}

// Get returns the currently stored state
func (m *InMemoryState) Get() (MonitoredIPs, error) {
	return *m.stored, nil
}

// InSync tests whether the internal state corresponds to the provided one
func (m *InMemoryState) InSync(with MonitoredIPs) bool {
	return reflect.DeepEqual(with, *m.stored)
}

// Reset returns the state to its default value
func (m *InMemoryState) Reset() error {
	m.stored = &MonitoredIPs{}
	return nil
}
