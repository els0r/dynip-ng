package state

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// File supplies methods to handle the state via a file
type File struct {
	fd   *os.File
	path string
}

// NewFile creates a new File to store the state in
func NewFile(path string) *File {
	return &File{path: path}
}

// Set writes a the state to disk in YAML representation
func (f *File) Set(ips MonitoredIPs) error {
	if f.open(os.O_WRONLY|os.O_CREATE|os.O_TRUNC) != nil {
		return fmt.Errorf("unable to open state file")
	}
	defer f.close()
	return yaml.NewEncoder(f.fd).Encode(&ips)
}

// Get reads the state from a YAML file from disk
func (f *File) Get() (MonitoredIPs, error) {
	stored := MonitoredIPs{}
	if f.open(os.O_RDONLY) != nil {
		return stored, fmt.Errorf("unable to open state file")
	}
	defer f.close()

	err := yaml.NewDecoder(f.fd).Decode(&stored)
	if err != nil {
		return MonitoredIPs{}, err
	}
	return stored, nil
}

// Reset deletes the state file
func (f *File) Reset() error {
	// close file in case it hasn't been closed
	f.close()
	return os.Remove(f.path)
}

func (f *File) open(mode int) error {
	var err error
	f.fd, err = os.OpenFile(f.path, mode, 0666)
	return err
}

func (f *File) close() error {
	if f.fd == nil {
		return nil
	}
	return f.fd.Close()
}
