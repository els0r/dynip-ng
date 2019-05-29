// Package listener monitors IP changes on a network interface
package listener

import (
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/els0r/dynip-ng/pkg/cfg"
	"github.com/els0r/dynip-ng/pkg/update"
	logger "github.com/els0r/log"
)

var log, _ = logger.NewFromString("console", logger.WithLevel(logger.DEBUG))

const defaultDiskLocation = "/var/run/.cf-dyn-ip"

// StoredIPs stores the observed IP addresses on disk
type StoredIPs struct {
	IPv4 string
	IPv6 string
}

type ipState struct {
	diskLocation string

	fromStorage StoredIPs
	fromIface   StoredIPs
}

// state loading and writing
func (s *ipState) write(dst io.Writer) error {
	return yaml.NewEncoder(dst).Encode(&s.fromIface)
}

func (s *ipState) writeToDisk() error {
	fd, err := os.Create(s.diskLocation)
	if err != nil {
		return err
	}
	defer fd.Close()
	return s.write(fd)
}

func (s *ipState) loadFromDisk() error {
	fd, err := os.Open(s.diskLocation)
	if err != nil {
		return err
	}
	defer fd.Close()
	return s.load(fd)
}

func (s *ipState) load(src io.Reader) error {
	return yaml.NewDecoder(src).Decode(&s.fromStorage)
}

func (s *ipState) inSync() bool {
	return reflect.DeepEqual(s.fromIface, s.fromStorage)
}

func (l *Listener) update() {
	var (
		err error
		ip  net.IP
	)

	// reset state in case the error is non-nil upon function return
	defer func(err error) {
		// reset and return
		if err != nil {
			log.Errorf("update failed: %s", err)
			l.state.reset()
			return
		}

		// write state to disk
		l.state.fromStorage = l.state.fromIface
		l.state.writeToDisk()
	}(err)

	// get current ip addresses
	ip, err = getLocalAddress(l.cfg.Iface)
	if err != nil {
		return
	}

	// assign IPs to state
	if ip.To4() != nil {
		l.state.fromIface.IPv4 = ip.String()
	} else {
		l.state.fromIface.IPv6 = ip.String()
	}

	// check update trigger condition
	if !l.state.inSync() {
		log.Info("running IP change destination updates")

		for _, u := range l.updaters {
			// update the IPs at the destination
			err = u.Update(l.state.fromIface.IPv4, l.cfg)

			// all updates have to be successful
			if err != nil {
				return
			}
		}
		return
	}
	log.Debug("IPs are equal. Nothing to do")
}

func (s *ipState) reset() {
	s.fromIface = StoredIPs{}
}

func newState() *ipState {
	return &ipState{
		diskLocation: defaultDiskLocation,
		fromStorage:  StoredIPs{},
		fromIface:    StoredIPs{},
	}
}

type Listener struct {
	state *ipState
	cfg   *cfg.Config

	// units that will receive an update
	updaters []update.Updater
}

// New creates a new listener
func New(cfg *cfg.Config, upds ...update.Updater) (*Listener, error) {
	l := new(Listener)

	if cfg == nil {
		return nil, fmt.Errorf("cannot run with <nil> config")
	}
	l.cfg = cfg

	// create initial state
	l.state = newState()

	if cfg.StateFile != "" {
		l.state.diskLocation = cfg.StateFile
	}

	// opportunistically attempt to load the state
	err := l.state.loadFromDisk()
	if err != nil {
		log.Debugf("loading state failed: %s", err)
	}

	// assign updaters
	l.updaters = upds

	return l, nil
}

// Run starts the IP change listener
func (l *Listener) Run() chan struct{} {

	log.Debugf("Running with config: %s", l.cfg)

	// setup time interval for periodic checks
	ticker := time.NewTicker(time.Duration(l.cfg.Interval) * time.Minute)
	defer ticker.Stop()

	// on return, write the last state one more time
	defer l.state.writeToDisk()

	// check and update if necessary
	l.update()

	// go into monitoring mode
	stopChan := make(chan struct{})
	go func(stop chan struct{}) {
		for {
			select {
			case <-ticker.C:
				// check and update if necessary
				l.update()
			case <-stop:
				log.Info("stopped listening for IP updates")
				return
			}
		}
	}(stopChan)
	return stopChan
}

func getLocalAddress(iface string) (net.IP, error) {

	// get the interface
	ifi, err := net.InterfaceByName(iface)
	if err != nil {
		return nil, err
	}

	// get all addresses for the interface
	var addrs []net.Addr
	addrs, err = ifi.Addrs()
	if err != nil {
		return nil, err
	}

	// get IP address for interface
	for _, a := range addrs {
		switch v := a.(type) {
		case *net.IPAddr:
			return v.IP, nil
		case *net.IPNet:
			return v.IP, nil
		}
	}
	return nil, fmt.Errorf("no IP address found for interface %q", iface)
}
