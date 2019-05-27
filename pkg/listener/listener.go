// Package listener monitors IP changes on a network interface
package listener

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"reflect"
	"syscall"
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

	fromDisk  StoredIPs
	fromIface StoredIPs
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
	return yaml.NewDecoder(src).Decode(&s.fromDisk)
}

func (s *ipState) update(cfg *cfg.Config) error {
	// get current ip addresses
	ip, err := getLocalAddress(cfg.Iface)
	if err != nil {
		return err
	}

	// assign IPs to state
	if ip.To4() != nil {
		s.fromIface.IPv4 = ip.String()
	} else {
		s.fromIface.IPv6 = ip.String()
	}

	// check for equality (currently
	if !(reflect.DeepEqual(s.fromIface, s.fromDisk)) {
		log.Info("running cloudflare update")

		var u update.Updater

		// TODO move into listener constructor
		u, err = update.NewCloudFlareUpdate(cfg.API.Key, cfg.API.Email)
		if err != nil {
			return err
		}

		// update the IPs at the destination
		err = u.Update(s.fromIface.IPv4, cfg)
		if err != nil {
			return err
		}

		// write state to disk
		s.fromDisk = s.fromIface
		s.writeToDisk()
	} else {
		log.Debug("IPs are equal. Nothing to do")
	}
	return nil
}

func (s *ipState) reset() {
	s.fromIface = StoredIPs{}
}

// Run starts the IP change listener
func Run(cfg *cfg.Config) error {
	if cfg == nil {
		return fmt.Errorf("cannot run with nil config")
	}

	log.Debugf("Running with config: %s", cfg)

	// setup time interval for periodic checks
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Minute)
	defer ticker.Stop()

	// we quit on encountering SIGTERM or SIGINT
	sigExitChan := make(chan os.Signal, 1)
	signal.Notify(sigExitChan, syscall.SIGTERM, os.Interrupt)

	// create initial state
	state := &ipState{
		diskLocation: defaultDiskLocation,
		fromDisk:     StoredIPs{},
		fromIface:    StoredIPs{},
	}
	if cfg.StateFile != "" {
		state.diskLocation = cfg.StateFile
	}
	defer state.writeToDisk()

	// opportunistically attempt to load the state
	err := state.loadFromDisk()
	if err != nil {
		log.Debugf("loading state failed: %s", err)
	}

	// check and update if necessary
	err = state.update(cfg)
	if err != nil {
		log.Errorf("update failed: %s", err)
		state.reset()
	}

	// go into monitoring mode
	for {
		select {
		case <-ticker.C:
			// check and update if necessary
			err = state.update(cfg)
			if err != nil {
				log.Errorf("update failed: %s", err)

				// reset state
				state.reset()
			}
		case <-sigExitChan:
			log.Info("Shutting down")
			return nil
		}
	}
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
