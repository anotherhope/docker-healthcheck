package healthcheck

import (
	"context"
	"strings"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
)

const containerID = "CONTAINER ID"
const containerNames = "NAMES"
const containerImage = "IMAGE"
const containerStatus = "STATUS"
const containerHealthCheck = "HEALTHCHECK"

type Data struct {
	Lines  [][]string
	Length []int
}

func getCanonicalContainerName(Names []string) string {
	for _, name := range Names {
		if strings.LastIndex(name, "/") == 0 {
			return name[1:]
		}
	}

	return Names[0][1:]
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func Analyse(dockerCli command.Cli, targets []string, only bool) (*Data, error) {
	cli := dockerCli.Client()
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})

	if err != nil {
		return nil, err
	}

	data := &Data{
		Lines: [][]string{{
			containerID,
			containerNames,
			containerImage,
			containerStatus,
			containerHealthCheck,
		}},
		Length: []int{
			len(containerID),
			len(containerNames),
			len(containerImage),
			len(containerStatus),
			len(containerHealthCheck),
		},
	}

	for _, container := range containers {
		inpected, err := cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			return nil, err
		}

		id := container.ID[0:12]
		names := getCanonicalContainerName(container.Names)
		image := container.Image
		if strings.HasPrefix(image, "sha256:") {
			image = image[7:19]
		}
		status := inpected.State.Status
		healthcheck := "-"

		if inpected.State.Health != nil {
			healthcheck = inpected.State.Health.Status
		}

		if len(id) > data.Length[0] {
			data.Length[0] = len(id)
		}

		if len(names) > data.Length[1] {
			data.Length[1] = len(names)
		}

		if len(image) > data.Length[2] {
			data.Length[2] = len(image)
		}

		if len(status) > data.Length[3] {
			data.Length[3] = len(status)
		}

		if len(healthcheck) > data.Length[4] {
			data.Length[4] = len(healthcheck)
		}

		if only && healthcheck == "-" || !contains(targets, names) {
			continue
		}

		data.Lines = append(
			data.Lines,
			[]string{
				id,
				names,
				image,
				status,
				healthcheck,
			},
		)
	}

	return data, nil
}
