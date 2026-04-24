package update

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/els0r/dynip-ng/pkg/cfg"
)

func TestFormatEndpoint(t *testing.T) {
	cases := []struct {
		ip   string
		port int
		want string
		err  bool
	}{
		{"1.2.3.4", 51820, "1.2.3.4:51820", false},
		{"2001:db8::1", 51820, "[2001:db8::1]:51820", false},
		{"::1", 51820, "[::1]:51820", false},
		{"not-an-ip", 51820, "", true},
	}
	for _, c := range cases {
		got, err := formatEndpoint(c.ip, c.port)
		if (err != nil) != c.err {
			t.Fatalf("%s: err=%v want-err=%v", c.ip, err, c.err)
		}
		if got != c.want {
			t.Fatalf("%s: got %q want %q", c.ip, got, c.want)
		}
	}
}

func TestWireGuardUpdate_InvokesWgSet(t *testing.T) {
	var calls [][]string
	fake := func(_ context.Context, name string, args ...string) ([]byte, error) {
		calls = append(calls, append([]string{name}, args...))
		return nil, nil
	}

	c := &cfg.WireGuardConfig{
		Peers: []cfg.WireGuardPeer{
			{Interface: "wg0", PublicKey: "ABCDEF=", Port: 51820},
			{Interface: "wg1", PublicKey: "GHIJKL=", Port: 51821},
		},
	}
	w, err := NewWireGuardUpdate(c, WithRunner(fake))
	if err != nil {
		t.Fatal(err)
	}

	if err := w.Update(context.Background(), "203.0.113.5"); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 invocations, got %d", len(calls))
	}
	want := "wg set wg0 peer ABCDEF= endpoint 203.0.113.5:51820"
	if got := strings.Join(calls[0], " "); got != want {
		t.Fatalf("call[0] got %q want %q", got, want)
	}
}

func TestWireGuardUpdate_IPv6Brackets(t *testing.T) {
	var captured []string
	fake := func(_ context.Context, name string, args ...string) ([]byte, error) {
		captured = append([]string{name}, args...)
		return nil, nil
	}
	c := &cfg.WireGuardConfig{
		Peers: []cfg.WireGuardPeer{{Interface: "wg0", PublicKey: "K=", Port: 51820}},
	}
	w, _ := NewWireGuardUpdate(c, WithRunner(fake))
	if err := w.Update(context.Background(), "2001:db8::1"); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(captured, " ")
	if !strings.Contains(joined, "[2001:db8::1]:51820") {
		t.Fatalf("expected bracketed v6 endpoint, got %q", joined)
	}
}

func TestWireGuardUpdate_PropagatesError(t *testing.T) {
	fake := func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return []byte("permission denied"), fmt.Errorf("exit 1")
	}
	c := &cfg.WireGuardConfig{Peers: []cfg.WireGuardPeer{{Interface: "wg0", PublicKey: "K=", Port: 51820}}}
	w, _ := NewWireGuardUpdate(c, WithRunner(fake))
	err := w.Update(context.Background(), "1.2.3.4")
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected error carrying stderr, got %v", err)
	}
}
