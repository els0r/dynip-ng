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

	"github.com/els0r/dynip-ng/pkg/cfg"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Test the configuration file for errors",
	Long: `Test the configuration file for errors.

This will check if the YAML syntax is okay and whether
all parameters have been provided`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Validating configuration at: %s\n", cfgPath)

		config, err := cfg.ParseFile(cfgPath)
		if err != nil {
			return err
		}
		fmt.Printf("Config:\n\t%s", config)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
