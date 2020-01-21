package cmd

import (
	"github.com/spf13/cobra"
	"github.com/towoe/gclone/repo"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print the entries",
	Long: "Print the entries of the storage file. Does not check the" +
		" status.",
	Run: list,
}

func list(cmd *cobra.Command, args []string) {
	i, _ := cmd.Flags().GetString("index")
	r := repo.CurrentRegister(i)
	r.List()
}
