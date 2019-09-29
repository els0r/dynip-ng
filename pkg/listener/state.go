package listener

import (
	"fmt"
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

// String outputs the stored IPv4 and IPv6 information
func (m MonitoredIPs) String() string {
	var v4, v6 = m.IPv4, m.IPv6
	empty := "<EMPTY>"
	if v6 == "" {
		v6 = empty
	}
	if v4 == "" {
		v4 = empty
	}
	return fmt.Sprintf("v4=%s, v6=%s", v4, v6)
}

// Set writes a the state to disk in YAML representation
func (f *FileState) Set(ips MonitoredIPs) error {
	if f.open() != nil {
		return fmt.Errorf("unable to open state file")
	}
	defer f.close()
	return yaml.NewEncoder(f.fd).Encode(&ips)
}

// Get reads the state from a YAML file from disk
func (f *FileState) Get() (MonitoredIPs, error) {
	stored := MonitoredIPs{}
	if f.open() != nil {
		return stored, fmt.Errorf("unable to open state file")
	}
	defer f.close()

	err := yaml.NewDecoder(f.fd).Decode(&stored)
	if err != nil {
		return MonitoredIPs{}, err
	}
	return stored, nil
}

// InSync tests whether the internal state corresponds to the provided one
func (f *FileState) InSync(with MonitoredIPs) bool {
	stored, err := f.Get()
	if err != nil {
		return false
	}
	return reflect.DeepEqual(with, stored)
}

// Reset writes an epty state to disk
func (f *FileState) Reset() error {
	return f.Set(MonitoredIPs{})
}

// FileState supplies methods to handle the state via a file
type FileState struct {
	fd   *os.File
	path string
}

// NewFileState creates a new FileState
func NewFileState(path string) (*FileState, error) {
	f := new(FileState)
	f.path = path

	return f, nil
}

func (f *FileState) open() error {
	var err error
	f.fd, err = os.OpenFile(f.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	return err
}

func (f *FileState) close() error {
	if f.fd == nil {
		return nil
	}
	return f.fd.Close()
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
