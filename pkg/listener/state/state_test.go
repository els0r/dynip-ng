package state

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/els0r/dynip-ng/pkg/cfg"
)

var (
	testDataDir   = "testdata"
	testFileState = ".test.state"
)

func TestMain(m *testing.M) {
	// make sure all file artefacts are removed
	err := os.MkdirAll(testDataDir, 0777)
	if err != nil {
		fmt.Fprintf(os.Stderr, "test directory creation failed: %s", err)
		os.Exit(1)
	}

	code := m.Run()

	err = os.RemoveAll(testDataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "test directory deletion failed: %s", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func TestState(t *testing.T) {

	memory, _ := New(&cfg.StateConfig{Type: "memory"})
	state, err := New(&cfg.StateConfig{
		Type: "file", Location: filepath.Join(testDataDir, testFileState),
	})
	if err != nil {
		t.Fatalf("could not create file state: %s", err)
	}

	_, err = New(nil)
	if err == nil {
		t.Fatalf("should have failed on nil state config")
	}

	_, err = New(&cfg.StateConfig{Type: "unknown"})
	if err == nil {
		t.Fatalf("should have failed on invalid state config")
	}

	var tests = []struct {
		name  string
		state State
		ips   MonitoredIPs
	}{
		{
			"in memory state",
			memory,
			MonitoredIPs{IPv4: "192.168.1.1"},
		},
		{
			"file state",
			state,
			MonitoredIPs{IPv4: "192.168.1.1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// get new state
			ims := test.state

			// set it
			ips := test.ips

			// states shouldn't be equal
			var ipsGot MonitoredIPs
			ipsGot, _ = ims.Get()
			if Equal(ipsGot, ips) {
				t.Fatalf("unitialized state shouldn't be in sync with non-empty IPs")
			}

			// set the ips
			err := ims.Set(ips)
			if err != nil {
				t.Fatalf("failed to set state: %s", err)
			}

			ipsGot, err = ims.Get()
			if err != nil {
				t.Fatalf("failed to get state: %s", err)
			}
			if !Equal(ipsGot, ips) {
				t.Fatalf("retrieved state and input IPs should be identical")
			}

			// change the monitored IP
			ips.IPv4 = "172.16.0.1"
			ipsGot, err = ims.Get()
			if err != nil {
				t.Fatalf("failed to get state: %s", err)
			}
			if Equal(ipsGot, ips) {
				t.Fatalf("input IPs changed and cannot be equal to internal state")
			}

			// reset state
			err = ims.Reset()
			if err != nil {
				t.Fatalf("failed to reset state: %s", err)
			}

			// states shouldn't be equal
			ipsGot, _ = ims.Get()
			if Equal(ipsGot, ips) {
				t.Fatalf("reset state shouldn't be in sync with non-empty IPs")
			}
		})
	}
}
