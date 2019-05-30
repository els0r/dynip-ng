package cfg

import (
	"strings"
	"testing"
)

var tests = []struct {
	name       string
	shouldPass bool
	cfg        string
}{
	{
		"valid configuration",
		true,
		`state_file: "/root/.ip-state"
cloudflare:
    access:
        key: 123
        email: test@example.com
    zone: example.ch
    record: dynip
listen:
    interval: 10
    iface: eth0
        `,
	},
	{
		"wrong interval value",
		false,
		`state_file: "/root/.ip-state"
cloudflare:
    access:
        key: 123
        email: test@example.com
    zone: example.ch
    record: dynip
listen:
    interval: -1
    iface: eth0
        `,
	},
	{
		"zone missing",
		false,
		`state_file: "/root/.ip-state"
cloudflare:
    access:
        key: 123
        email: test@example.com
    record: dynip
listen:
    interval: -1
    iface: eth0
        `,
	},
	{
		"faulty YAML",
		false,
		`zone: example.ch
-
        `,
	},
	{
		"invalid API configuration - key missing",
		false,
		`state_file: "/root/.ip-state"
cloudflare:
    access:
        email: test@example.com
    record: dynip
listen:
    interval: -1
    iface: eth0
        `,
	},
}

func TestValidate(t *testing.T) {

	// run tests
	for i, test := range tests {
		// run each case as a sub test
		t.Run(test.name, func(t *testing.T) {
			// create reader to parse config
			r := strings.NewReader(test.cfg)

			// parse config
			cfg, err := Parse(r)
			if test.shouldPass {
				if err != nil {
					t.Fatalf("[%d] couldn't parse config: %s", i, err)
				}
				t.Log(cfg)
			} else {
				if err == nil {
					t.Log(cfg)
					t.Fatalf("[%d] config parsing should have failed but didn't", i)
				}
				t.Logf("[%d] provoked expected error: %s", i, err)
				return
			}
		})
	}
}
