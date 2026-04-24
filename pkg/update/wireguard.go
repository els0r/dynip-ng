package update

import (
	"context"
	"fmt"
	"net"
	"os/exec"

	"github.com/els0r/dynip-ng/pkg/cfg"
	"github.com/els0r/dynip-ng/pkg/logging"
	log "github.com/els0r/log"
)

// WireGuardUpdate re-resolves the endpoint of configured WireGuard peers by
// invoking `wg set <iface> peer <pubkey> endpoint <ip>:<port>`. The kernel
// caches the original endpoint address per peer, so a DNS change alone is not
// enough to steer the tunnel at a new public IP.
type WireGuardUpdate struct {
	peers  []cfg.WireGuardPeer
	wgBin  string
	runner commandRunner
	log    log.Logger
}

// commandRunner is an indirection around exec.CommandContext so the updater
// can be unit-tested without a real wg binary.
type commandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

func defaultRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// WGOption configures the WireGuard updater.
type WGOption func(*WireGuardUpdate)

// WithRunner swaps out the command runner. Intended for tests.
func WithRunner(r commandRunner) WGOption {
	return func(w *WireGuardUpdate) { w.runner = r }
}

// WithWGBinary overrides the wg binary path (default: "wg" resolved via PATH).
func WithWGBinary(path string) WGOption {
	return func(w *WireGuardUpdate) { w.wgBin = path }
}

// NewWireGuardUpdate constructs a new WireGuard endpoint updater.
func NewWireGuardUpdate(c *cfg.WireGuardConfig, opts ...WGOption) (*WireGuardUpdate, error) {
	if c == nil {
		return nil, fmt.Errorf("wireguard: nil config")
	}
	w := &WireGuardUpdate{
		peers:  c.Peers,
		wgBin:  "wg",
		runner: defaultRunner,
		log:    logging.Get(),
	}
	for _, o := range opts {
		o(w)
	}
	return w, nil
}

// Name returns a human-readable identifier for the updater.
func (w *WireGuardUpdate) Name() string { return "wireguard updater" }

// formatEndpoint builds the "host:port" string passed to `wg set ... endpoint`.
// wg(8) requires IPv6 literals to be wrapped in square brackets.
func formatEndpoint(ip string, port int) (string, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", fmt.Errorf("not a valid IP: %q", ip)
	}
	if parsed.To4() == nil {
		return fmt.Sprintf("[%s]:%d", parsed.String(), port), nil
	}
	return fmt.Sprintf("%s:%d", parsed.String(), port), nil
}

// Update re-resolves the endpoint of every configured WireGuard peer to the new IP.
func (w *WireGuardUpdate) Update(ctx context.Context, ip string) error {
	if ip == "" {
		return fmt.Errorf("wireguard: empty IP")
	}
	for _, p := range w.peers {
		endpoint, err := formatEndpoint(ip, p.Port)
		if err != nil {
			return err
		}
		args := []string{"set", p.Interface, "peer", p.PublicKey, "endpoint", endpoint}
		w.log.Debugf("running %s %v", w.wgBin, args)
		out, err := w.runner(ctx, w.wgBin, args...)
		if err != nil {
			return fmt.Errorf("wg set %s peer: %w (output=%q)", p.Interface, err, string(out))
		}
	}
	return nil
}
