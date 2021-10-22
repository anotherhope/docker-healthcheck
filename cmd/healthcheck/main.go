package main

import (
	"fmt"
	"os"
	"time"

	"github.com/anotherhope/docker-healthcheck/internal/healthcheck"
	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/nsf/termbox-go"
	"github.com/spf13/cobra"
)

var filter []string
var only bool
var waithc string
var wait bool
var timeout time.Duration

func main() {
	plugin.Run(func(dockerCli command.Cli) *cobra.Command {
		cmd := &cobra.Command{
			Short:   "Display healthcheck on container",
			Example: "docker healthcheck --wait --filter CONTAINER_NAMES,CONTAINER_NAMES --only --timeout 10",
			Args: func(cmd *cobra.Command, args []string) error {
				if waithc != "" && !(waithc == "starting" || waithc == "healthy" || waithc == "unhealthy") {
					return fmt.Errorf("docker: --waithc args is not valid (starting|healthy|unhealthy)")
				}

				return nil
			},
			Use: "healthcheck",
			RunE: func(cmd *cobra.Command, args []string) error {
				hc, err := healthcheck.Analyse(dockerCli)
				if err != nil {
					return err
				}

				if only {
					hc = hc.Only()
				}

				if len(filter) > 0 {
					hc, err = hc.Targets(filter)
					if err != nil {
						return err
					}
				}

				if wait || waithc != "" {
					err := termbox.Init()
					if err != nil {
						return err
					}

					exit := make(chan os.Signal, 1)

					go func() {
						for {
							switch ev := termbox.PollEvent(); ev.Key {
							case termbox.KeyCtrlC:
								exit <- os.Interrupt
							}
						}
					}()

					go func() {
						for {
							for y, info := range hc.GetInfos() {
								for x, c := range info.Print(hc.GetMeta()) {
									termbox.SetCell(x, y, c, termbox.ColorWhite, termbox.ColorDefault)
								}
							}
							termbox.Flush()
							time.Sleep(1 * time.Second)
						}
					}()
					<-exit
					termbox.Close()
				} else {
					for _, info := range hc.GetInfos() {
						fmt.Println(info.Print(hc.GetMeta()))
					}
				}

				return nil
			},
		}

		cmd.Flags().StringVar(&waithc, "healthcheck", "", "Wait for all targeted containers obtained (starting|healthy|unhealthy) healthcheck")
		cmd.Flags().BoolVar(&wait, "wait", false, "Wait for all targeted containers")
		cmd.Flags().BoolVar(&only, "only", false, "Display only available healthcheck")
		cmd.Flags().StringSliceVarP(&filter, "filter", "f", nil, "Reduce containers to observe")
		cmd.Flags().DurationVarP(&timeout, "timeout", "t", 0, "Set timeout")

		if timeout > 0 {
			time.AfterFunc(timeout*time.Second, func() {
				if wait {
					termbox.Close()
				}
				fmt.Println("docker: timeout exceeded")
				os.Exit(1)
			})
		}

		return cmd
	}, manager.Metadata{
		SchemaVersion: "0.1.0",
		Vendor:        "Docker Inc.",
		Version:       "testing",
		Experimental:  true,
	})
}
