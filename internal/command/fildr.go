package command

import (
	"context"
	"fildr-cli/internal/log"
	runner2 "fildr-cli/internal/runner"
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog"
	golog "log"
	"os"
	"os/signal"
	"strings"
)

func newFildrCmd(version string, gitCommit string, buildTime string) *cobra.Command {
	fildrCmd := &cobra.Command{
		Use:   "fildr",
		Short: "ops client",
		Long:  "pusher is a client for prometheus gateway",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if err := bindViper(cmd); err != nil {
				golog.Println("unable to bind flags: ", err)
			}

			logLevel := 0
			if viper.GetBool("verbose") {
				logLevel = 1
			}

			z, err := log.Init(logLevel)
			if err != nil {
				golog.Printf("unable to initialize logger: %v", err)
				os.Exit(1)
			}
			defer func() {
				_ = z.Sync()
			}()

			logger := log.Wrap(z.Sugar())

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt)

			runCh := make(chan bool, 1)
			shutdownCh := make(chan bool, 1)

			logger.Infof("fildr-cli started.")

			go func() {

				klogVerbosity := viper.GetString("klog-verbosity")
				var klogOpts []string

				klogFlagSet := flag.NewFlagSet("klog", flag.ContinueOnError)
				if klogVerbosity == "" {
					klogOpts = append(klogOpts,
						fmt.Sprintf("-logtostderr=false"),
						fmt.Sprintf("-alsologtostderr=false"),
					)
				} else {
					klogOpts = append(klogOpts,
						fmt.Sprintf("-v=%s", klogVerbosity),
						fmt.Sprintf("-logtostderr=true"),
						fmt.Sprintf("-alsologtostderr=true"),
					)
				}

				klog.InitFlags(klogFlagSet)
				_ = klogFlagSet.Parse(klogOpts)

				options := runner2.Options{}

				runner, err := runner2.NewRunner(ctx, logger, options)
				if err != nil {
					golog.Println("unable to start runner: ", err)
					os.Exit(1)
				}

				runner.Start(ctx, nil, shutdownCh)

				runCh <- true
			}()

			select {
			case <-sigCh:
				logger.Infof("Shutting fildr down due to interrupt")
				cancel()
				<-shutdownCh
			case <-runCh:
				logger.Infof("Fildr has exited")
			}
		},
	}

	fildrCmd.Flags().SortFlags = false

	fildrCmd.Flags().StringP("context", "", "", "initial context ")

	return fildrCmd
}

func bindViper(cmd *cobra.Command) error {
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("FILDR")
	viper.AutomaticEnv()

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	if err := viper.BindEnv("fildr-config-home", "FILDR_CONFIG_HOME"); err != nil {
		return err
	}

	if err := viper.BindEnv("home", "HOME"); err != nil {
		return err
	}

	return nil
}
