// Package listener monitors IP changes on a network interface
package listener

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/els0r/dynip-ng/pkg/cfg"
	"github.com/els0r/dynip-ng/pkg/listener/state"
	"github.com/els0r/dynip-ng/pkg/logging"
	"github.com/els0r/dynip-ng/pkg/update"
	log "github.com/els0r/log"
	"github.com/miekg/dns"
)

const defaultUpdateTimeout = 30 * time.Second

func (l *Listener) update() {
	var (
		err error
		ip  net.IP
	)

	ctx, cancel := context.WithTimeout(context.Background(), defaultUpdateTimeout)
	defer cancel()

	// reset state in case the error is non-nil upon function return
	defer func(err error) {
		// reset and return
		if err != nil {
			serr := l.state.Reset()
			if serr != nil {
				l.log.Warnf("failed to reset state: %s", serr)
			}
			return
		}
	}(err)

	// get current ip addresses
	if !l.cfg.IsLAN {
		ip, err = getLocalAddress(l.cfg.Iface)
		if err != nil {
			l.log.Errorf("failed to get IP address on %q: %s", l.cfg.Iface, err)
			return
		}
	} else {
		l.log.Infof("getting external IP address on %q (lan=%v)", l.cfg.Iface, l.cfg.IsLAN)
		// get the IP address via a call to dyn
		ip, err = getExternalIPAddress(ctx)
		if err != nil {
			l.log.Errorf("failed to get external IP address on %q: %s", l.cfg.Iface, err)
			return
		}
	}
	l.log.Debugf("current interface IP is %q", ip)

	// assign read out IPs to state
	var ips = state.NewMonitoredIPs(ip)

	// get stored state
	storedIPs, err := l.state.Get()
	if err != nil {
		l.log.Warnf("failed to get state: %s", err)
	}

	// check update trigger condition
	if !state.Equal(storedIPs, ips) {
		l.log.Infof("IP(s) changed (%s): running destination updates", ips)

		tstart := time.Now()
		var numErrors int
		for _, u := range l.updaters {

			l.log.Debugf("running %s", u.Name())

			// update the IPs at the destination
			err = u.Update(ctx, ips.IPv4)

			// log errors
			if err != nil {
				l.log.Errorf("%s: %s", u.Name(), err)
				numErrors++
			}
		}
		if numErrors == 0 {
			l.log.Infof("all destinations updated in %s", time.Now().Sub(tstart))

			// write state only if all updates were successful
			err = l.state.Set(ips)
			if err != nil {
				l.log.Warnf("failed to set new state: %s", err)
			}
		} else if numErrors == len(l.updaters) {
			l.log.Errorf("all destinations encountered update errors. Time elapsed: %s", time.Now().Sub(tstart))
		} else {
			l.log.Warnf("some destinations encountered update errors. Time elapsed: %s", time.Now().Sub(tstart))
		}
		return
	}
	l.log.Debug("IPs are equal. Nothing to do")
}

// Listener listens for IP changes on an interface and updates all its configured destinations
type Listener struct {
	state state.State
	cfg   *cfg.ListenConfig

	// units that will receive an update
	updaters []update.Updater

	// logger for injection
	log log.Logger
}

// New creates a new listener
func New(cfg *cfg.ListenConfig, state state.State, upds ...update.Updater) (*Listener, error) {
	l := new(Listener)

	// get the program level logger
	l.log = logging.Get()

	if cfg == nil {
		return nil, fmt.Errorf("cannot run without listener config")
	}
	l.cfg = cfg

	// create initial state
	l.state = state

	// opportunistically attempt to load the state
	_, err := l.state.Get()
	if err != nil {
		l.log.Debugf("loading state failed: %s", err)
		l.state.Reset()
	}

	// assign updaters
	l.updaters = upds

	return l, nil
}

// Run starts the IP change listener
func (l *Listener) Run() chan struct{} {

	l.log.Debugf("running with config: %s", l.cfg)

	// setup time interval for periodic checks
	ticker := time.NewTicker(time.Duration(l.cfg.Interval) * time.Minute)

	// check and update if necessary
	l.log.Debug("running initial IP update check")
	l.update()

	// go into monitoring mode
	stopChan := make(chan struct{})
	go func(stop chan struct{}) {
		for {
			select {
			case <-ticker.C:
				// check and update if necessary
				l.log.Debug("running periodic IP update check")
				l.update()
			case <-stop:
				l.log.Info("stopped listening for IP updates")

				ticker.Stop()
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

const (
	openDNSResolver = "resolver1.opendns.com."
	extIPHostname   = "myip.opendns.com."
)

// getExternalIPAddress looks up the public IP address behind which the dynip service runs
func getExternalIPAddress(ctx context.Context) (net.IP, error) {
	// get IP of open DNS resolver
	resolverIPs, err := net.LookupIP(openDNSResolver)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve IP of OpenDNS resolver: %w", err)
	}
	if len(resolverIPs) == 0 {
		return nil, fmt.Errorf("no OpenDNS resolver IP returned", err)
	}
	resolverIP := resolverIPs[0]

	msg := new(dns.Msg)
	msg.SetQuestion(extIPHostname, dns.TypeA)
	c := new(dns.Client)

	reply, _, err := c.ExchangeContext(ctx, msg, resolverIP.String()+":53")
	if err != nil {
		return nil, err
	}
	if len(reply.Answer) == 0 {
		return nil, fmt.Errorf("empty resource record returned")
	}
	r := reply.Answer[0]
	return r.(*dns.A).A, nil
}
