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
	"github.com/spf13/cobra"
	"gitlab.jule.lan/infra/dynip-ng/pkg/cfg"
	"gitlab.jule.lan/infra/dynip-ng/pkg/listener"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the IP update listener",
	Long: `Listens for changes on interface and updates configured A record in
Cloudflare`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := cfg.ParseFile(cfgPath)
		if err != nil {
			return err
		}
		return listener.Run(config)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
