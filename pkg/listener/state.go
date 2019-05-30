package listener

import (
	"os"
	"reflect"

	yaml "gopkg.in/yaml.v2"
)

type State interface {
	Setter
	Getter
	Reset() error
	InSync(MonitoredIPs) bool
}

type Setter interface {
	Set(MonitoredIPs) error
}

type Getter interface {
	Get() (MonitoredIPs, error)
}

// MonitoredIPs stores the IP objects
type MonitoredIPs struct {
	IPv4 string
	IPv6 string
}

// state loading and writing
func (f *FileState) Set(ips MonitoredIPs) error {
	f.stored = &ips
	return yaml.NewEncoder(f.fd).Encode(f.stored)
}

func (f *FileState) Get() (MonitoredIPs, error) {
	err := yaml.NewDecoder(f.fd).Decode(f.stored)
	return *f.stored, err
}

func (f *FileState) InSync(with MonitoredIPs) bool {
	return reflect.DeepEqual(with, *f.stored)
}

func (f *FileState) Reset() error {
	return f.Set(MonitoredIPs{})
}

type FileState struct {
	fd     *os.File
	stored *MonitoredIPs
}

func NewFileState(path string) (*FileState, error) {
	var err error

	f := new(FileState)

	f.fd, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type InMemoryState struct {
	stored *MonitoredIPs
}

func NewInMemoryState() *InMemoryState {
	return &InMemoryState{
		stored: &MonitoredIPs{},
	}
}

func (m *InMemoryState) Set(ips MonitoredIPs) error {
	m.stored = &ips
	return nil
}

func (m *InMemoryState) Get() (MonitoredIPs, error) {
	return *m.stored, nil
}

func (m *InMemoryState) InSync(with MonitoredIPs) bool {
	return reflect.DeepEqual(with, *m.stored)
}

func (m *InMemoryState) Reset() error {
	m.stored = &MonitoredIPs{}
	return nil
}
