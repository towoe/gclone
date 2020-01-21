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
	Short: "Clone a repository",
	Long: "Clone a repository and add it to the storage file.\n" +
		"Following the git convention the second argument will be" +
		" the name of the new directory.\n" +
		"For more complicated clones, use `git clone` followed by an" +
		" `gclone add`.",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		i, _ := cmd.Flags().GetString("index")
		r := repo.CurrentRegister(i)
		r.LoadRemotes()
		dest := ""
		if len(args) >= 2 {
			dest = args[1]
		}
		r.Clone(args[0], dest)
	},
}
