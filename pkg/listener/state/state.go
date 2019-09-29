package state

import (
	"fmt"
	"net"
	"strings"

	"github.com/els0r/dynip-ng/pkg/cfg"
)

// State stores and informs about the current state of the monitored interface's IPs
type State interface {
	Setter
	Getter
	Reset() error
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

// NewMonitoredIPs creates a new container for the changed IP based on the
// interface reading
func NewMonitoredIPs(ip net.IP) MonitoredIPs {
	var ips = MonitoredIPs{}
	if ip.To4() != nil {
		ips.IPv4 = ip.String()
	} else {
		ips.IPv6 = ip.String()
	}
	return ips
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

// Equal checks if IPs a are identical to ips b
func Equal(a, b MonitoredIPs) bool {
	return a.IPv4 == b.IPv4 && a.IPv6 == b.IPv6
}

// New returns a state implementation based on the provided type
func New(config *cfg.StateConfig) (State, error) {
	if config == nil {
		return nil, fmt.Errorf("no state config provided")
	}

	t := strings.ToLower(config.Type)
	switch t {
	case "memory":
		return NewInMemory(), nil
	case "file":
		return NewFile(config.Location), nil
	}
	return nil, fmt.Errorf("state type %q not (yet) supported", config.Type)
}
