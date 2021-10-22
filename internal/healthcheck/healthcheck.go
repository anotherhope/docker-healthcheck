package healthcheck

import (
	"context"
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const pad int = 3

type HealthCheck struct {
	cli   client.APIClient
	infos []*Info
	meta  Meta
}

func (hc *HealthCheck) Only() *HealthCheck {
	infos := []*Info{}

	for _, info := range hc.infos {
		if info.healthCheck != "-" {
			infos = append(infos, info)
		}
	}

	hcn := &HealthCheck{
		cli:   hc.cli,
		infos: infos,
	}

	return hcn.updateMeta()
}

func (hc *HealthCheck) Targets(targets []string) (*HealthCheck, error) {
	filter := targets
	for i, info := range hc.infos {
		index := find(filter, info.names)
		if index != -1 && i != 0 {
			filter = append(filter[:index], filter[index+1:]...)
		}
	}

	if len(filter) > 0 {
		return nil, fmt.Errorf("docker: Error targets containers %v not found", filter)
	}

	infos := []*Info{}

	for i, info := range hc.infos {
		if contains(targets, info.names) || i == 0 {
			infos = append(infos, info)
		}
	}

	hcn := &HealthCheck{
		cli:   hc.cli,
		infos: infos,
	}

	return hcn.updateMeta(), nil
}

func find(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return -1
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func Analyse(dockerCli command.Cli) (*HealthCheck, error) {
	hc := &HealthCheck{
		cli: dockerCli.Client(),
		infos: []*Info{{
			id:          "CONTAINER ID",
			names:       "NAMES",
			image:       "IMAGE",
			status:      "STATUS",
			healthCheck: "HEALTHCHECK",
		}},
	}

	return hc.refreshData()
}

func (hc *HealthCheck) refreshData() (*HealthCheck, error) {
	containers, err := hc.cli.ContainerList(context.Background(), types.ContainerListOptions{})

	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		inpected, err := hc.cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			return nil, err
		}

		ir := &InfoRaw{
			id:          container.ID,
			names:       container.Names,
			image:       container.Image,
			status:      inpected.State.Status,
			healthCheck: inpected.State.Health,
		}

		is := ir.ToString()
		hc.infos = append(hc.infos, is)
	}

	return hc.updateMeta(), nil
}

func (hc *HealthCheck) updateMeta() *HealthCheck {

	for _, i := range hc.infos {
		if hc.meta.id < pad+len(i.id) {
			hc.meta.id = pad + len(i.id)
		}

		if hc.meta.names < pad+len(i.names) {
			hc.meta.names = pad + len(i.names)
		}

		if hc.meta.image < pad+len(i.image) {
			hc.meta.image = pad + len(i.image)
		}

		if hc.meta.status < pad+len(i.status) {
			hc.meta.status = pad + len(i.status)
		}

		if hc.meta.healthCheck < pad+len(i.healthCheck) {
			hc.meta.healthCheck = pad + len(i.healthCheck)
		}
	}

	return hc
}

func (hc *HealthCheck) GetInfos() []*Info {
	return hc.infos
}

func (hc *HealthCheck) GetMeta() Meta {
	return hc.meta
}

/*

type Data struct {
	Lines  [][]string
	Length []int
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

		if (only && healthcheck == "-") || len(targets) > 0 && !contains(targets, names) {
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
*/
