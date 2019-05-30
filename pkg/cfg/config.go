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
	// Listen configures where to listen for IP updates
	Listen *ListenConfig

	// StateFile stores the location of the state file
	StateFile string `yaml:"state_file"`

	// API configures the accessto cloudflare
	Cloudflare *CloudflareAPI
}

// ListenConfig configures the listener
type ListenConfig struct {
	// External interface to monitor changes on
	Iface string

	// Interval stores the time between periodic checks
	Interval int
}

// CloudflareAPI configures the accessto cloudflare
type CloudflareAPI struct {
	Access struct {
		// Key is the API key for Cloudflare
		Key string

		// Email is the email associated with the API key
		Email string
	}

	// Record which should be changed
	Record string

	// Zone holding the record
	Zone string
}

func (c *CloudflareAPI) validate() error {
	if c.Access.Key == "" {
		return fmt.Errorf("cloudflare: no API key provided")
	}
	if c.Access.Email == "" {
		return fmt.Errorf("cloudflare: no API email provided")
	}
	if c.Zone == "" {
		return fmt.Errorf("cloudflare: no zone to update record in provided")
	}
	return nil
}

// New creates a default configuration
func New() *Config {
	return &Config{
		Listen: &ListenConfig{
			Iface:    "eth0", // assumes that eth0 is the default interface
			Interval: 5,      // standard check is every 5 minutes
		},
	}
}

// String provides quick info about what this configuration updates
func (c *Config) String() string {
	return fmt.Sprintf("Updates every: %dm; Iface: %q",
		c.Listen.Interval,
		c.Listen.Iface)
}

// the validator interface is a contract to show if a concrete type is
// configured according to its predefined value range
type validator interface {
	validate() error
}

func (l *ListenConfig) validate() error {
	if l.Iface == "" {
		return fmt.Errorf("listener: no interface provided on which daemon monitors changes")
	}
	if l.Interval <= 0 {
		return fmt.Errorf("listener: checking period must be greater zero (minutes)")
	}
	return nil
}

func (c *Config) validate() error {
	if c.Listen == nil {
		return fmt.Errorf("no listener configuration provided")
	}
	if c.Cloudflare == nil {
		return fmt.Errorf("no cloudflare configuration provided")
	}
	return nil
}

// Validate validates the configuration file
func (c *Config) Validate() error {
	// run all config subsection validators. Order matters here
	for _, section := range []validator{
		c,
		c.Listen,
		c.Cloudflare,
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
