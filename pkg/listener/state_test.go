package listener

import (
	"reflect"
	"testing"
)

func TestState(t *testing.T) {
	var tests = []struct {
		name  string
		state State
		ips   MonitoredIPs
	}{
		{
			"in memory state",
			NewInMemoryState(),
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
			if ims.InSync(ips) {
				t.Fatalf("unitialized state shouldn't be in sync with non-empty IPs")
			}

			// set the ips
			err := ims.Set(ips)
			if err != nil {
				t.Fatalf("failed to set state: %s", err)
			}
			if !ims.InSync(ips) {
				t.Fatalf("state and input IPs should be identical")
			}

			var ipsGot MonitoredIPs
			ipsGot, err = ims.Get()
			if err != nil {
				t.Fatalf("failed to get state: %s", err)
			}
			if !reflect.DeepEqual(ipsGot, ips) {
				t.Fatalf("retrieved state and input IPs should be identical")
			}

			// reset state
			err = ims.Reset()
			if err != nil {
				t.Fatalf("failed to reset state: %s", err)
			}

			// states shouldn't be equal
			if ims.InSync(ips) {
				t.Fatalf("reset state shouldn't be in sync with non-empty IPs")
			}
		})
	}
}
