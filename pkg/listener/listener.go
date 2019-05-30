// Package listener monitors IP changes on a network interface
package listener

import (
	"fmt"
	"net"
	"time"

	"github.com/els0r/dynip-ng/pkg/cfg"
	"github.com/els0r/dynip-ng/pkg/update"
	logger "github.com/els0r/log"
)

var log, _ = logger.NewFromString("console", logger.WithLevel(logger.DEBUG))

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
			l.state.Reset()
			return
		}
	}(err)

	// get current ip addresses
	ip, err = getLocalAddress(l.cfg.Iface)
	if err != nil {
		return
	}

	// assign IPs to state
	var ips = MonitoredIPs{}
	if ip.To4() != nil {
		ips.IPv4 = ip.String()
	} else {
		ips.IPv6 = ip.String()
	}

	// check update trigger condition
	if !l.state.InSync(ips) {
		log.Info("running IP change destination updates")

		for _, u := range l.updaters {
			// update the IPs at the destination
			err = u.Update(ips.IPv4, l.cfg)

			// all updates have to be successful
			if err != nil {
				return
			}
		}
		log.Info("all destinations updated")

		// write state to disk
		l.state.Set(ips)
		return
	}
	log.Debug("IPs are equal. Nothing to do")
}

type Listener struct {
	state State
	cfg   *cfg.Config

	// units that will receive an update
	updaters []update.Updater
}

// New creates a new listener
func New(cfg *cfg.Config, state State, upds ...update.Updater) (*Listener, error) {
	l := new(Listener)

	if cfg == nil {
		return nil, fmt.Errorf("cannot run with <nil> config")
	}
	l.cfg = cfg

	// create initial state
	l.state = state

	// opportunistically attempt to load the state
	_, err := l.state.Get()
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
