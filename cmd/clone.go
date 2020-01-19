package cmd

import (
	"github.com/spf13/cobra"
	"github.com/towoe/gclone/repo"
)

func init() {
	rootCmd.AddCommand(cloneCmd)
}

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone the specified arguemnt.",
	Long:  "Clone the give path and add it to the storage file. An additional path can be used to specify the destination.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		i, _ := cmd.Flags().GetString("index")
		r := repo.CurrentRegister(i)
		r.LoadRemotes()
		r.Clone(args[0], args[1])
	},
}
