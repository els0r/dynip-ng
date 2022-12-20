// Package update is responsible for updating destinations using IP
package update

import "context"

// Updater is an interface that takes a configuration and updates the IP
type Updater interface {
	Update(ctx context.Context, IP string) error
	Name() string
}
