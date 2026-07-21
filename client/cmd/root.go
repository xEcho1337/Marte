package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "marte",
	Aliases: []string{"mrt"},
	Short:   "Client CLI for Marte",
	Long:    "Marte: CLI client to manage environments and exploits during Attack/Defense CTFs.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
