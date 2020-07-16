package command

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/user"
)

var cfg = `
[gateway]
url = "https://api.fildr.com/fildr-miner"
token = ""
instance = ""
evaluation = 5
`

func newInitializationCmd(version, gitCommit, buildTime string) *cobra.Command {

	initializationCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialization config",
		Long:  "Initialization config",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			user, err := user.Current()
			path := user.HomeDir + "/.fildr"
			_, err = os.Stat(path)
			if err != nil {
				err = os.Mkdir(path, os.ModePerm)
				if err != nil {
					fmt.Fprintf(out, "Error creating folder: %s", err.Error())
					os.Exit(1)
				}
			}

			ioutil.WriteFile(path+"/config.toml", []byte(cfg), os.ModePerm)
			fmt.Fprintln(out, "Initialization complete")
		},
	}
	return initializationCmd
}
