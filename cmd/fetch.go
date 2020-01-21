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
	Short: "Download updates",
	Long: "Download updates for each entry of the storage file.\n" +
		"This will only get the objects from the main tracked remote " +
		" repository.\n" +
		"Additional arguments are forwarded to the invocation of" +
		" `git fetch`, except the help flag.",
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
			if strings.Compare("-h", a) == 0 || strings.Compare("--help", a) == 0 {
				cmd.Help()
				return
			}
			argsNoHelp = append(argsNoHelp, a)
		}
		r.Fetch(strings.Join(argsNoHelp, " "))
	},
}
