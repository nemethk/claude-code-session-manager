package cmd

import (
	"github.com/nemethk/claude-code-session-manager/internal/upgrade"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade ccsm to the latest version",
	Long: `Download and install the latest version of ccsm.

Requires sudo to replace the binary at /usr/local/bin/ccsm.

Usage:
  sudo ccsm upgrade

This command:
  1. Fetches the latest release from GitHub
  2. Skips download if already on the latest version
  3. Downloads the binary for your OS and architecture
  4. Installs it to /usr/local/bin/ccsm
  5. Verifies the installation`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return upgrade.Run(rootCmd.Version)
	},
}
