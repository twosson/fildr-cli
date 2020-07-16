package command

import (
	"fildr-cli/internal/config"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func newInitializationCmd() *cobra.Command {

	initializationCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialization config",
		Long:  "Initialization config",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			if err := config.InitializationConfig(); err != nil {
				fmt.Fprintln(out, "initialization config err: ", err)
				os.Exit(1)
			}
			fmt.Fprintln(out, "initialization complete.")
		},
	}
	return initializationCmd
}
