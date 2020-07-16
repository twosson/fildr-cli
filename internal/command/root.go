package command

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func Execute(version string, gitCommit string, buildTime string) {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	rootCmd := newRoot(version, gitCommit, buildTime)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRoot(version string, gitCommit string, buildTime string) *cobra.Command {
	rootCmd := newFildrCmd(version, gitCommit, buildTime)
	rootCmd.AddCommand(newVersionCmd(version, gitCommit, buildTime))
	rootCmd.AddCommand(newInitializationCmd(version, gitCommit, buildTime))
	return rootCmd
}
