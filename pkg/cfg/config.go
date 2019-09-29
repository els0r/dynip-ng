// Package cfg configures the dyn-ip daemon
package cfg

import (
	"fmt"
	"io"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Config holds the dyn-ip configuration
type Config struct {
	// Listen configures where to listen for IP updates
	Listen *ListenConfig

	// StateFile stores the location of the state file
	State *StateConfig

	// Destinations stores all places to be updated
	Destinations *DestinationsConfig

	// Logging configuration
	Logging *LoggingConfig `yaml:"logging"`
}

// StateConfig configures how the state should be kept and where
type StateConfig struct {
	// Type of state tracking (file, memory, etc.)
	Type string

	// Location of the state resource
	Location string
}

func (s *StateConfig) validate() error {
	t := strings.ToLower(s.Type)

	switch t {
	case "memory", "file":
		break
	default:
		return fmt.Errorf("state type %q is not (yet) supported", s.Type)
	}
	if s.Location == "" && t != "memory" {
		return fmt.Errorf("state location not provided")
	}
	return nil
}

// LoggingConfig can reconfigure the program logger
type LoggingConfig struct {
	// Where are logs written
	Destination string `yaml:"destination"`

	// Log level
	Level string `yaml:"level"`
}

// DestinationsConfig stores all output destinations
type DestinationsConfig struct {
	// configures the cloudflare API
	Cloudflare *CloudflareAPI `yaml:"cloudflare,omitempty"`
	// configures the file update config
	File *FileConfig `yaml:"file,omitempty"`
}

// FileConfig stores parameters for
type FileConfig struct {
	Template string `yaml:"template"`
	Output   string `yaml:"output"`
}

func (f *FileConfig) validate() error {
	if f.Template == "" {
		return fmt.Errorf("file: no input template provided")
	}
	if f.Output == "" {
		return fmt.Errorf("file: no output file provided")
	}
	return nil
}

func (d DestinationsConfig) validate() error {
	var sections []validator

	// check if there is at least one destination configured
	if d.Cloudflare != nil {
		sections = append(sections, d.Cloudflare)
	}
	if d.File != nil {
		sections = append(sections, d.File)
	}
	if len(sections) == 0 {
		return fmt.Errorf("no destination for IP provided. Need at least one")
	}

	// run all config subsection validators. Order matters here
	for _, section := range sections {
		err := section.validate()
		if err != nil {
			return err
		}
	}
	return nil
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

	// list of Zones to update
	Zones map[string]*Zone
}

// Zone stores the DNS objects that should be updated
type Zone struct {
	// Record to change
	Record string
}

func (z *Zone) validate() error {
	return nil
}

func (c *CloudflareAPI) validate() error {
	if c.Access.Key == "" {
		return fmt.Errorf("cloudflare: no API key provided")
	}
	if c.Access.Email == "" {
		return fmt.Errorf("cloudflare: no API email provided")
	}
	if len(c.Zones) == 0 {
		return fmt.Errorf("cloudflare: no zone to update record in provided")
	}
	for name, zone := range c.Zones {
		if name == "" {
			return fmt.Errorf("cloudflare: zone with no name provided")
		}
		err := zone.validate()
		if err != nil {
			return err
		}
	}
	return nil
}

// New creates a default configuration
func New() *Config {
	return &Config{
		State: &StateConfig{
			Type: "memory", // by default, track the state in memory
		},
		Listen: &ListenConfig{
			Interval: 5, // standard check is every 5 minutes
		},
		Logging: &LoggingConfig{
			Destination: "console", // log to console by default
			Level:       "INFO",
		},
	}
}

// String provides quick info about what this configuration updates
func (c *Config) String() string {
	return c.Listen.String()
}

// String provides quick info about what the listener does
func (l *ListenConfig) String() string {
	return fmt.Sprintf("updates every: %dm; Iface: %q",
		l.Interval,
		l.Iface,
	)
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
	if c.Destinations == nil {
		return fmt.Errorf("no destination configuration provided")
	}
	if c.State == nil {
		return fmt.Errorf("no state configuration provided")
	}
	return nil
}

// Validate validates the configuration file
func (c *Config) Validate() error {
	// run all config subsection validators. Order matters here
	for _, section := range []validator{
		c,
		c.Listen,
		c.State,
		c.Destinations,
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
