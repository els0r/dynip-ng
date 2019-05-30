// Package update is responsible or updating DNS records
package update

// Updater is an interface that takes a configuration and updates the IP
type Updater interface {
	Update(IP string) error
}
