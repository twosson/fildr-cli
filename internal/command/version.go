package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

func newVersionCmd(version, gitCommit, buildTime string) *cobra.Command {

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Long:  "Version for octant binary",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()

			fmt.Fprintln(out, "Version: ", version)
			fmt.Fprintln(out, "Git commit: ", gitCommit)
			fmt.Fprintln(out, "Built: ", buildTime)
		},
	}
	return versionCmd
}
