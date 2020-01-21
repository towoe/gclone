package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/towoe/gclone/repo"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add directories to the storage",
	Long: "Add directories to the storage file by appending one or more\n" +
		"paths to a git directory as an argument.",
	Run: func(cmd *cobra.Command, args []string) {
		i, _ := cmd.Flags().GetString("index")
		r := repo.CurrentRegister(i)
		r.LoadRemotes()
		fmt.Println(args)
		for _, d := range args {
			r.Add(d)
		}
	},
}
