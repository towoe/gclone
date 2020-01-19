package cmd

import (
	"github.com/spf13/cobra"
)

var (
	indexFile string
	rootCmd   = &cobra.Command{
		Use: "gclone",
		Run: status,
	}
)

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&indexFile, "index", "i", "", "index file (default is $XDG_DATA_HOME/gclone/register.json)")
}
