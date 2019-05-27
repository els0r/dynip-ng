// Package update is responsible or updating DNS records
package update

import "github.com/els0r/dynip-ng/pkg/cfg"

// Updater is an interface that takes a configuration and updates the IP
type Updater interface {
	Update(IP string, cfg *cfg.Config) error
}
