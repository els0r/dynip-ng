package update

import (
	"fmt"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/els0r/dynip-ng/pkg/cfg"

	log "github.com/els0r/log"
)

var logger, _ = log.NewFromString("console", log.WithLevel(log.DEBUG))

// CloudFlareUpdate communicates with the cloudflare API to change records
type CloudFlareUpdate struct {
	api CloudflareAPI
	cfg *cfg.CloudflareAPI
}

// CloudflareAPI allows us to decouple the third-party CloudFlare API implementation.
// This is useful for mocking the cloudflare api for testing.
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
func NewCloudFlareUpdate(cfg *cfg.CloudflareAPI, opts ...CFOption) (*CloudFlareUpdate, error) {
	c := new(CloudFlareUpdate)

	// store zone and record update config
	c.cfg = cfg

	// Construct a new API object
	var err error
	c.api, err = cloudflare.New(cfg.Access.Key, cfg.Access.Email)
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
func (c *CloudFlareUpdate) Update(IP string) error {

	logger.Debug("updating Cloudflare zones")

	for name, zoneCfg := range c.cfg.Zones {
		logger.Debugf("updating Cloudflare zone: %s", name)

		// Fetch the zone ID
		zoneID, err := c.api.ZoneIDByName(name)
		if err != nil {
			return err
		}

		// Fetch all records for a zone
		recs, err := c.api.DNSRecords(zoneID, cloudflare.DNSRecord{})
		if err != nil {
			return err
		}

		recordToUpdate := name
		if zoneCfg.Record != "" {
			recordToUpdate = zoneCfg.Record + "." + name
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
					}
					logger.Debugf("updated A record '%s' with IP address '%s'", recordToUpdate, IP)
					return nil
				}
			}
		}
		return fmt.Errorf("record %q was not found", recordToUpdate)
	}
	return nil
}
