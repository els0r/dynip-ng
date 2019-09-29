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
		"valid configuration (both destinations)",
		true,
		`---
state_file: "/root/.ip-state"
destinations:
    cloudflare:
        access:
            key: 123
            email: test@example.com

        zones:
            example.ch:
                record: dynip
    file:
        template: /path/to/template
        output: /path/to/output

listen:
    interval: 10
    iface: eth0
        `,
	},
	{
		"no listen interface",
		false,
		`---
state_file: "/root/.ip-state"
destinations:
    cloudflare:
        access:
            key: 123
            email: test@example.com

        zones:
            example.ch:
                record: dynip
    file:
        template: /path/to/template
        output: /path/to/output

listen:
    interval: 10
        `,
	},
	{
		"valid configuration (cloudflare)",
		true,
		`---
state_file: "/root/.ip-state"
destinations:
    cloudflare:
        access:
            key: 123
            email: test@example.com

        zones:
            example.ch:
                record: dynip

listen:
    interval: 10
    iface: eth0
        `,
	},
	{
		"no access section (cloudflare)",
		false,
		`---
state_file: "/root/.ip-state"
destinations:
    cloudflare:
        zones:
            example.ch:
                record: dynip

listen:
    interval: 10
    iface: eth0
        `,
	},
	{
		"no email in access (cloudflare)",
		false,
		`---
state_file: "/root/.ip-state"
destinations:
    cloudflare:
				access:
						key: 123
        zones:
            example.ch:
                record: dynip

listen:
    interval: 10
    iface: eth0
        `,
	},
	{
		"empty zone (cloudflare)",
		false,
		`---
state_file: "/root/.ip-state"
destinations:
    cloudflare:
				access:
						key: 123
        zones:
            example.ch:
                record: ""

listen:
    interval: 10
    iface: eth0
        `,
	},
	{
		"no destinations",
		false,
		`---
state_file: "/root/.ip-state"
listen:
    interval: 10
    iface: eth0
        `,
	},
	{
		"valid configuration (file)",
		true,
		`state_file: "/root/.ip-state"
destinations:
    file:
        template: /path/to/template
        output: /path/to/output
listen:
    interval: 10
    iface: eth0
        `,
	},
	{
		"wrong interval value",
		false,
		`state_file: "/root/.ip-state"
destinations:
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
destinations:
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
destinations:
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
					t.Logf("config: %s", cfg)
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
