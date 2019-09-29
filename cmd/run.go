// Copyright © 2019 Lennart Elsen (lel)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/els0r/dynip-ng/pkg/cfg"
	"github.com/els0r/dynip-ng/pkg/listener"
	"github.com/els0r/dynip-ng/pkg/update"
	"github.com/spf13/cobra"
)

const defaultStateDiskLocation = "/var/run/.cf-dyn-ip"

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

		// create updaters
		var updaters []update.Updater
		dests := config.Destinations

		if dests.Cloudflare != nil {
			cu, err := update.NewCloudFlareUpdate(dests.Cloudflare)
			if err != nil {
				return err
			}
			updaters = append(updaters, cu)
		}
		if dests.File != nil {
			fu, err := update.NewFileUpdate(dests.File)
			if err != nil {
				return err
			}
			updaters = append(updaters, fu)
		}

		// check if the state file is set, otherwise take default path
		path := defaultStateDiskLocation
		if config.StateFile != "" {
			path = config.StateFile
		}

		state, err := listener.NewFileState(path)
		if err != nil {
			return fmt.Errorf("failed to create state: %s", err)
		}

		// create listener
		var l *listener.Listener
		l, err = listener.New(config.Listen, state, updaters...)
		if err != nil {
			return fmt.Errorf("failed to create listener: %s", err)
		}

		// and run it
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
