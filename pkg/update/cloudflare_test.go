package update

import (
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

func (m *mockAPI) DNSRecords(zoneID string, r cloudflare.DNSRecord) ([]cloudflare.DNSRecord, error) {
	if zoneID != m.zoneID {
		return nil, fmt.Errorf("no records found for zone ID: %s", zoneID)
	}
	return m.records, nil
}

func (m *mockAPI) UpdateDNSRecord(zoneID string, recordID string, r cloudflare.DNSRecord) error {
	if zoneID != m.zoneID {
		return fmt.Errorf("zone ID %s not found", zoneID)
	}
	for i, record := range m.records {
		if record.ID == recordID {
			m.records[i] = r
			return nil
		}
	}
	return fmt.Errorf("record %q could not be found in zone", recordID)
}

func TestNewCloudFlare(t *testing.T) {

	var tests = []struct {
		name       string
		key        string
		email      string
		shouldPass bool
	}{
		{"init cloudflare updater", "api_test_key_1", "test@example.com", true},
		{"no key", "", "test@example.com", false},
		{"no email", "", "test@example.com", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create new object
			c, err := NewCloudFlareUpdate(test.key, test.email)
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
		cfg        *cfg.Config
		shouldPass bool
	}{
		{"record found", IP, &cfg.Config{Zone: zoneName, Record: recordName}, true},
		{"zone not found", IP, &cfg.Config{Zone: "notAvailable", Record: recordName}, false},
		{"record not found", IP, &cfg.Config{Zone: zoneName, Record: "notAvailable"}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create new updater with mock API
			c, err := NewCloudFlareUpdate("unused", "unused", WithCFAPI(&mockAPI{
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
			err = c.Update(test.IP, test.cfg)
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
