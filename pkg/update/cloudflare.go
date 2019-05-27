package update

import (
	"fmt"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/els0r/dynip-ng/pkg/cfg"

	logger "github.com/els0r/log"
)

var log, _ = logger.NewFromString("console", logger.WithLevel(logger.DEBUG))

// CloudFlareUpdate communicates with the cloudflare API to change records
type CloudFlareUpdate struct {
	api CloudflareAPI
}

// api allows us to mock the cloudflare api for testing
type CloudflareAPI interface {
	ZoneIDByName(string) (string, error)
	DNSRecords(string, cloudflare.DNSRecord) ([]cloudflare.DNSRecord, error)
	UpdateDNSRecord(string, string, cloudflare.DNSRecord) error
}

// CFOption allows to modify the cloudflare updater
type CFOption func(c *CloudFlareUpdate)

// WithCFAPI allows to pass another API than the default one
func WithCFAPI(api CloudflareAPI) CFOption {
	return func(c *CloudFlareUpdate) {
		c.api = api
	}
}

// NewCloudFlareUpdate return a new cloudflare updater
func NewCloudFlareUpdate(key, email string, opts ...CFOption) (*CloudFlareUpdate, error) {
	c := new(CloudFlareUpdate)

	// Construct a new API object
	var err error
	c.api, err = cloudflare.New(key, email)
	if err != nil {
		return nil, err
	}

	// apply functional options
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Update changes the record from the config in Cloudflare to `ip`
func (c *CloudFlareUpdate) Update(IP string, cfg *cfg.Config) error {

	// Fetch the zone ID
	zoneID, err := c.api.ZoneIDByName(cfg.Zone)
	if err != nil {
		return err
	}

	// Fetch all records for a zone
	recs, err := c.api.DNSRecords(zoneID, cloudflare.DNSRecord{})
	if err != nil {
		return err
	}

	recordToUpdate := cfg.Zone
	if cfg.Record != "" {
		recordToUpdate = cfg.Record + "." + cfg.Zone
	}

	for _, r := range recs {
		// only take the A record
		if r.Type == "A" {

			// update if it is the IP address record
			if r.Name == recordToUpdate {

				// set to new IP address
				r.Content = IP

				err = c.api.UpdateDNSRecord(zoneID, r.ID, r)
				if err != nil {
					return err
				} else {
					log.Infof("Updated A record '%s' with IP address '%s'", recordToUpdate, IP)
					return nil
				}
			}
		}
	}
	return fmt.Errorf("record %q was not found", recordToUpdate)
}
