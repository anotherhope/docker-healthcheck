package main

import (
	"fmt"
	"time"

	"github.com/anotherhope/docker-healthcheck/internal/healthcheck"
	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/nsf/termbox-go"
	"github.com/spf13/cobra"
)

var only bool
var wait bool
var filter []string
var timeout int
var healthcheckArgs string
var exit chan error

func init() {
	exit = make(chan error, 1)
}

func main() {

	plugin.Run(func(dockerCli command.Cli) *cobra.Command {
		cmd := &cobra.Command{
			Short:   "Display healthcheck of all container",
			Example: "docker healthcheck --f CONTAINER_NAMES,CONTAINER_NAMES healthy",
			Use:     "healthcheck starting | healthy | unhealthy",
			Args:    cobra.MaximumNArgs(1),
			PreRunE: func(cmd *cobra.Command, args []string) error {
				if len(args) > 0 {
					healthcheckArgs = args[0]
					if !(healthcheckArgs == "starting" || healthcheckArgs == "healthy" || healthcheckArgs == "unhealthy") {
						return fmt.Errorf("docker: args is not valid ( starting | healthy | unhealthy )")
					}
				}
				return nil
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				if timeout > 0 {
					time.AfterFunc(time.Duration(timeout)*time.Second, func() {
						exit <- fmt.Errorf("docker: timeout exceeded")
					})
				}

				hc := healthcheck.Make(dockerCli)

				if only || len(filter) > 0 {
					hc.SetOnly()
				}

				if len(filter) > 0 {
					hc.SetTargets(filter, wait)
				}

				if err := hc.RefreshData(); err != nil {
					return err
				}

				infos := hc.GetInfos()
				if !wait {
					for _, info := range infos {
						fmt.Println(info.Print(hc.GetMeta()))
					}
					return nil
				}

				if len(infos)-1 == len(filter) {
					if len(args) == 0 || hc.IsValid(args[0]) {
						termbox.Close()
						return nil
					}
				}

				if err := termbox.Init(); err != nil {
					return err
				}

				defer termbox.Close()

				go func() {
					for {
						switch ev := termbox.PollEvent(); ev.Key {
						case termbox.KeyCtrlC:
							exit <- nil
						}
					}
				}()

				go func() {
					for {
						for y, info := range infos {
							for x, c := range info.Print(hc.GetMeta()) {
								termbox.SetCell(x, y, c, termbox.ColorWhite, termbox.ColorDefault)
							}
						}
						termbox.Flush()
						termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)
						time.Sleep(time.Second)

						if err := hc.RefreshData(); err != nil {
							exit <- err
						}

						infos = hc.GetInfos()

						if len(infos)-1 == len(filter) {
							if len(args) == 0 || hc.IsValid(args[0]) {
								exit <- nil
							}
						}

						if len(args) > 0 && hc.IsValid(args[0]) {
							exit <- nil
						}
					}
				}()

				return <-exit
			},
		}

		cmd.Flags().BoolVar(&wait, "wait", false, "Wait for all targeted containers")
		cmd.Flags().BoolVar(&only, "only", false, "Display only available healthcheck")
		cmd.Flags().StringSliceVarP(&filter, "filter", "f", nil, "Reduce containers to observe by name")
		cmd.Flags().IntVarP(&timeout, "timeout", "t", 0, "Set timeout")

		return cmd
	}, manager.Metadata{
		SchemaVersion: "0.1.0",
		Vendor:        "Docker Inc.",
		Version:       "testing",
		Experimental:  true,
	})
}
