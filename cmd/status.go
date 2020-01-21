package cmd

import (
	"github.com/spf13/cobra"
	"github.com/towoe/gclone/repo"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the entries statuses",
	Long: "Print the status for each entry of the storage file.\n" +
		"By default the list is sorted by directories.\n",
	Run: status,
}

func init() {
	statusCmd.Flags().StringP("list", "l", "dir", "List the output based on [dir]ectory or [remote]")
}

func status(cmd *cobra.Command, args []string) {
	i, _ := cmd.Flags().GetString("index")
	r := repo.CurrentRegister(i)
	r.LoadRemotes()
	l, _ := cmd.Flags().GetString("list")
	r.Status(l)
	r.RemoveInvalidEntries(repo.DeleteAsk)
	r.Store()
}
