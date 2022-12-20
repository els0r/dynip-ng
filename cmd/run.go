package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/els0r/dynip-ng/pkg/cfg"
	"github.com/els0r/dynip-ng/pkg/listener"
	"github.com/els0r/dynip-ng/pkg/listener/state"
	"github.com/els0r/dynip-ng/pkg/logging"
	"github.com/els0r/dynip-ng/pkg/update"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the IP update listener",
	Long: `Listens for changes on interface and updates it's configured receivers
attributes. For example the A record on Cloudflare.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		// we quit on encountering SIGTERM or SIGINT
		sigExitChan := make(chan os.Signal, 1)
		signal.Notify(sigExitChan, syscall.SIGTERM, os.Interrupt)

		// parse config
		config, err := cfg.ParseFile(cfgPath)
		if err != nil {
			return err
		}

		// initialize logger
		err = logging.Init(config.Logging)
		if err != nil {
			return err
		}
		logging.Get().Debug("Initialized logger")

		// create updaters
		var updaters []update.Updater
		dests := config.Destinations

		if dests.Cloudflare != nil {
			cu, err := update.NewCloudFlareUpdate(dests.Cloudflare)
			if err != nil {
				return err
			}
			updaters = append(updaters, cu)
			logging.Get().Debug("Initialized cloudflare updates")
		}
		if dests.File != nil {
			fu, err := update.NewFileUpdate(dests.File)
			if err != nil {
				return err
			}
			updaters = append(updaters, fu)
			logging.Get().Debug("Initialized file updates")
		}

		// prepare the state
		state, err := state.New(config.State)
		if err != nil {
			return fmt.Errorf("failed to create state: %s", err)
		}
		logging.Get().Debug("Initialized state tracking")

		// create listener
		var l *listener.Listener
		l, err = listener.New(config.Listen, state, updaters...)
		if err != nil {
			return fmt.Errorf("failed to create listener: %s", err)
		}

		// and run it
		logging.Get().Debug("Spawning listener")
		stop := l.Run()

		// listen for the exit signal and stop the listener
		<-sigExitChan
		stop <- struct{}{}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
