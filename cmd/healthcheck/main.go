package main

import (
	"fmt"
	"strconv"

	"github.com/anotherhope/docker-healthcheck/internal/healthcheck"
	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"
)

var targets []string
var only bool

func main() {
	plugin.Run(func(dockerCli command.Cli) *cobra.Command {
		cmd := &cobra.Command{
			Use: "healthcheck ",
			RunE: func(cmd *cobra.Command, args []string) error {

				data, err := healthcheck.Analyse(dockerCli, targets, only)
				if err != nil {
					return err
				}

				for _, line := range data.Lines {
					fmt.Printf("%-"+strconv.Itoa(data.Length[0]+3)+"s"+
						"%-"+strconv.Itoa(data.Length[1]+3)+"s"+
						"%-"+strconv.Itoa(data.Length[2]+3)+"s"+
						"%-"+strconv.Itoa(data.Length[3]+3)+"s"+
						"%-"+strconv.Itoa(data.Length[4]+3)+"s"+
						"\n", line[0], line[1], line[2], line[3], line[4],
					)
				}

				return nil
			},
		}

		cmd.Flags().BoolVar(&only, "only", false, "Display only active healthcheck")
		cmd.Flags().StringSliceVarP(&targets, "for", "f", nil, "Provide filter targets")

		return cmd
	}, manager.Metadata{
		SchemaVersion: "0.1.0",
		Vendor:        "Docker Inc.",
		Version:       "testing",
		Experimental:  true,
	})
}
