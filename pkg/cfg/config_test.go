package cfg

import (
	"strings"
	"testing"
)

var validStateConfig = `
state:
        type: file
        location: "/root/.ip-state"
`

var invalidStateConfigs = []string{`
state:
        type: unknown
        location: unknown
`,
	`
state:
        type: file
`}

var validListenConfig = `
listen:
    interval: 10
    iface: eth0
`

var tests = []struct {
	name       string
	shouldPass bool
	cfg        string
}{
	{
		"valid configuration (both destinations)",
		true,
		`---` + validStateConfig + `
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
        ` + validListenConfig,
	},
	{
		"unsupported state type",
		false,
		`---` + invalidStateConfigs[0] + `
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
        ` + validListenConfig,
	},
	{
		"empty state location",
		false,
		`---` + invalidStateConfigs[1] + `
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
        ` + validListenConfig,
	},
	{
		"no listen interface",
		false,
		`---` + validStateConfig + `
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
		`---` + validStateConfig + `
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
		`---` + validStateConfig + `
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
		`---` + validStateConfig + `
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
		`---` + validStateConfig + `
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
		`---` + validStateConfig + `
listen:
    interval: 10
    iface: eth0
        `,
	},
	{
		"valid configuration (file)",
		true,
		`---` + validStateConfig + `
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
		`---` + validStateConfig + `
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
		`---` + validStateConfig + `
destinations:
    cloudflare:
        access:
            key: 123
            email: test@example.com

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
		`---` + validStateConfig + `
destinations:
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
