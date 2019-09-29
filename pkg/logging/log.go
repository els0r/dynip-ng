package logging

import (
	"github.com/els0r/dynip-ng/pkg/cfg"
	log "github.com/els0r/log"
)

// package level logger. Once set by Init(), it is not supposed to be modified
var logger log.Logger = log.NewDevNullLogger()

// Init initializes the program-wide logger
func Init(config *cfg.LoggingConfig) error {
	var err error
	logger, err = log.NewFromString(
		config.Destination,
		log.WithLevel(log.GetLevel(config.Level)),
	)
	return err
}

// Get returns the program-wide logger
func Get() log.Logger {
	return logger
}
