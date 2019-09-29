// Package update is responsible for updating destinations using IP
package update

// Updater is an interface that takes a configuration and updates the IP
type Updater interface {
	Update(IP string) error
	Name() string
}
