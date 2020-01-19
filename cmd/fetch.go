package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/towoe/gclone/repo"
)

func init() {
	rootCmd.AddCommand(fetchCmd)
}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Perform a fetch for each entry in the storage file.",
	Long:  "For each entry a 'git fetch' is executed. Additional arguments are forwarded to 'git fetch'",
	// TODO: disable parsing of additional arguments
	// Could be done with fetchCmd.Flags().SetInterspersed(false)
	DisableFlagParsing: true, // Forward the options as flags to the following fetches
	Run: func(cmd *cobra.Command, args []string) {
		i, _ := cmd.Flags().GetString("index")
		r := repo.CurrentRegister(i)
		r.LoadRemotes()
		var argsNoHelp []string
		// TODO remove this hack after parsing stops for non declared arguments
		for _, a := range args {
			if !(strings.Compare("-h", a) == 0 || strings.Compare("--help", a) == 0) {
				argsNoHelp = append(argsNoHelp, a)
			}
		}
		r.Fetch(strings.Join(argsNoHelp, " "))
	},
}
