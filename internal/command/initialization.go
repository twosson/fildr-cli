package command

import (
	"fildr-cli/internal/config"
	"fmt"
	"github.com/spf13/cobra"
	golog "log"
	"os"
	"time"
)

func newInitializationCmd() *cobra.Command {

	initializationCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialization config",
		Long:  "Initialization config",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()

			if err := bindViper(cmd); err != nil {
				golog.Println("unable to bind flags: ", err)
			}

			if err := config.InitializationConfig(); err != nil {
				fmt.Fprintln(out, "initialization config err: ", err)
				os.Exit(1)
			}
			fmt.Fprintln(out, "initialization complete.")
		},
	}

	initializationCmd.Flags().SortFlags = false
	initializationCmd.Flags().StringP("gateway.token", "", "", "config gateway token")
	initializationCmd.Flags().StringP("gateway.instance", "", "", "config gateway instance")
	initializationCmd.Flags().DurationP("gateway.evaluation", "", time.Second*5, "config gateway evaluation")
	initializationCmd.Flags().StringP("gateway.url", "", "https://api.fildr.com/fildr-miner", "config gateway url")

	return initializationCmd
}
