package update

import (
	"context"
	"fmt"
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/els0r/dynip-ng/pkg/cfg"
)

// mock API for cloudflare to test program flow
type mockAPI struct {
	zoneName string
	zoneID   string
	records  []cloudflare.DNSRecord
}

func (m *mockAPI) ZoneIDByName(name string) (string, error) {
	if name != m.zoneName {
		return "", fmt.Errorf("no zone ID found for name: %s", name)
	}
	return m.zoneID, nil
}

func (m *mockAPI) ListDNSRecords(_ context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, *cloudflare.ResultInfo, error) {
	if rc.Identifier != m.zoneID {
		return nil, nil, fmt.Errorf("no records found for zone ID: %s", rc.Identifier)
	}
	return m.records, &cloudflare.ResultInfo{}, nil
}

func (m *mockAPI) UpdateDNSRecord(_ context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
	if rc.Identifier != m.zoneID {
		return cloudflare.DNSRecord{}, fmt.Errorf("zone ID %s not found", rc.Identifier)
	}
	for i, record := range m.records {
		if record.ID == params.ID {
			m.records[i].Content = params.Content
			return m.records[i], nil
		}
	}
	return cloudflare.DNSRecord{}, fmt.Errorf("record %q could not be found in zone", params.ID)
}

func TestNewCloudFlare(t *testing.T) {

	var tests = []struct {
		name       string
		cfg        *cfg.CloudflareAPI
		shouldPass bool
	}{
		{
			"init cloudflare updater",
			&cfg.CloudflareAPI{
				Access: struct {
					Token string
					Key   string
					Email string
				}{"", "api_test_key_1", "test@example.com"},
			},
			true,
		},
		{
			"init cloudflare updater with token",
			&cfg.CloudflareAPI{
				Access: struct {
					Token string
					Key   string
					Email string
				}{"api_test_token_1", "", ""},
			},
			true,
		},
		{
			"no key",
			&cfg.CloudflareAPI{
				Access: struct {
					Token string
					Key   string
					Email string
				}{"", "", "test@example.com"},
			},
			false,
		},
		{
			"no email",
			&cfg.CloudflareAPI{
				Access: struct {
					Token string
					Key   string
					Email string
				}{"", "api_test_key_1", ""},
			},
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create new object
			c, err := NewCloudFlareUpdate(test.cfg)
			if test.shouldPass {
				if err != nil {
					t.Fatalf("couldn't create cloudflare updater: %s", err)
				}
			} else {
				if err == nil {
					t.Log(c)
					t.Fatalf("cloudflare updater instantiation should have failed but didn't")
				}
				t.Logf("provoked expected error: %s", err)
				return
			}
		})
	}
}

func TestCloudFlareUpdate(t *testing.T) {

	// mock API parameters
	var (
		zoneName, zoneID, recordID, recordName, IP = "testZone", "testZoneID", "testRecordID", "testRecordName", "192.168.1.1"
	)

	var tests = []struct {
		name       string
		IP         string
		cfg        *cfg.CloudflareAPI
		shouldPass bool
	}{
		{"record found", IP, &cfg.CloudflareAPI{
			Zones: map[string]*cfg.Zone{
				zoneName: &cfg.Zone{Record: recordName},
			},
			Access: struct{ Token, Key, Email string }{"", "key", "e@mail.com"},
		}, true},
		{"records found", IP, &cfg.CloudflareAPI{
			Zones: map[string]*cfg.Zone{
				zoneName: &cfg.Zone{Record: recordName},
			},
			Access: struct{ Token, Key, Email string }{"", "key", "e@mail.com"},
		}, true},
		{"zone not found", IP, &cfg.CloudflareAPI{
			Zones: map[string]*cfg.Zone{
				"notAvailable": &cfg.Zone{Record: recordName},
			},
			Access: struct{ Token, Key, Email string }{"", "key", "e@mail.com"},
		}, false},
		{"record not found", IP, &cfg.CloudflareAPI{
			Zones: map[string]*cfg.Zone{
				zoneName: &cfg.Zone{Record: "notAvailable"},
			},
			Access: struct{ Token, Key, Email string }{"", "key", "e@mail.com"},
		}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create new updater with mock API
			// TODO: rewrite the mock API to be able to test updates to multiple zones
			c, err := NewCloudFlareUpdate(test.cfg, WithCFAPI(&mockAPI{
				zoneName: zoneName,
				zoneID:   zoneID,
				records: []cloudflare.DNSRecord{
					cloudflare.DNSRecord{
						Type:    "A",
						ID:      recordID,
						Content: IP,
						Name:    recordName + "." + zoneName,
					},
				},
			}))
			if err != nil {
				t.Fatalf("couldn't create cloudflare updater: %s", err)
			}

			// update the record
			err = c.Update(context.Background(), test.IP)
			if test.shouldPass {
				if err != nil {
					t.Fatalf("cloudflare update failed: %s", err)
				}
			} else {
				if err == nil {
					t.Log(c)
					t.Fatalf("cloudflare update should have failed but didn't")
				}
				t.Logf("provoked expected error: %s", err)
				return
			}
		})
	}
}
