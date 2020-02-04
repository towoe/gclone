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
	statusCmd.Flags().StringP("sort", "s", "key", "Sort the output by [key] or [status]")
	statusCmd.Flags().BoolP("reverse", "r", false, "Reverse the order of sorting")
}

func status(cmd *cobra.Command, args []string) {
	i, _ := cmd.Flags().GetString("index")
	r := repo.CurrentRegister(i)
	r.LoadRemotes()
	list, _ := cmd.Flags().GetString("list")
	sorted, _ := cmd.Flags().GetString("sort")
	reverse, _ := cmd.Flags().GetBool("reverse")
	r.Status(list, sorted, reverse)
	if r.RemoveInvalidEntries(repo.DeleteAsk) {
		r.Store()
	}
}
