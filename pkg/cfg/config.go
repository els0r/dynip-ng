// Package cfg configures the dyn-ip daemon
package cfg

import (
	"fmt"
	"io"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// Config holds the dyn-ip configuration
type Config struct {
	// Record which should be changed
	Record string

	// Zone holding the record
	Zone string

	// External interface to monitor changes on
	Iface string

	// Interval stores the time between periodic checks
	Interval int

	// StateFile stores the location of the state file
	StateFile string `yaml:"state_file"`

	// API configures the accessto cloudflare
	API *APIConfig `yaml:"cloudflare_api"`
}

// APIConfig configures the accessto cloudflare
type APIConfig struct {
	// Key is the API key for Cloudflare
	Key string

	// Email is the email associated with the API key
	Email string
}

func (a *APIConfig) validate() error {
	if a.Key == "" {
		return fmt.Errorf("cloudflare_api: no API key provided")
	}
	if a.Email == "" {
		return fmt.Errorf("cloudflare_api: no API email provided")
	}
	return nil
}

// New creates a default configuration
func New() *Config {
	return &Config{
		Iface:    "eth0", // assumes that eth0 is the default interface
		Interval: 5,      // standard check is every 5 minutes
	}
}

// String provides quick info about what this configuration updates
func (c *Config) String() string {
	return fmt.Sprintf("Updates: %s.%s; Every: %dm; Iface: %q",
		c.Record, c.Zone,
		c.Interval,
		c.Iface)
}

// the validator interface is a contract to show if a concrete type is
// configured according to its predefined value range
type validator interface {
	validate() error
}

func (c *Config) validate() error {
	if c.Record == "" {
		return fmt.Errorf("no record to update provided")
	}
	if c.Zone == "" {
		return fmt.Errorf("no zone to update record in provided")
	}
	if c.Iface == "" {
		return fmt.Errorf("no interface provided on which daemon monitors changes")
	}
	if c.API == nil {
		return fmt.Errorf("no cloudflare configuration provided")
	}
	if c.Interval <= 0 {
		return fmt.Errorf("checking period must be greater zero (seconds)")
	}
	return nil
}

// Validate validates the configuration file
func (c *Config) Validate() error {
	// run all config subsection validators. Order matters here
	for _, section := range []validator{
		c,
		c.API,
	} {
		err := section.validate()
		if err != nil {
			return err
		}
	}
	return nil
}

// Parse parses the configuration from an io.Reader
func Parse(src io.Reader) (*Config, error) {
	c := New()

	err := yaml.NewDecoder(src).Decode(c)
	if err != nil {
		return nil, err
	}

	err = c.Validate()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ParseFile parses the configuration from a file
func ParseFile(path string) (*Config, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	return Parse(fd)
}
