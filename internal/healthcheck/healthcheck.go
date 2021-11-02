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
	cli     client.APIClient
	infos   []*Info
	meta    Meta
	targets []string
	wait    bool
	only    bool
}

func (hc *HealthCheck) SetOnly() {
	hc.only = true
}

func (hc *HealthCheck) SetTargets(targets []string, wait bool) {
	hc.targets = targets
	hc.wait = wait
}

func (hc *HealthCheck) filter() error {
	if len(hc.targets) > 0 {
		filter := make([]string, len(hc.targets))
		copy(filter, hc.targets)
		for i, info := range hc.infos {
			index := find(filter, info.names)
			if index != -1 && i != 0 {
				filter = append(filter[:index], filter[index+1:]...)
			}
		}
		if len(filter) > 0 && !hc.wait {
			hc.updateMeta()
			return fmt.Errorf("docker: Error targets containers %v not found", filter)
		}

		infos := []*Info{}

		for i, info := range hc.infos {
			if contains(hc.targets, info.names) || i == 0 {
				infos = append(infos, info)
			}
		}
		hc.infos = infos
	}

	hc.updateMeta()
	return nil
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

func Make(dockerCli command.Cli) *HealthCheck {
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

	return hc
}

func (hc *HealthCheck) RefreshData() error {
	containers, err := hc.cli.ContainerList(context.Background(), types.ContainerListOptions{})

	if err != nil {
		return err
	}

	hc.infos = hc.infos[:1]
	hc.meta = Meta{}

	for _, container := range containers {
		inpected, err := hc.cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			return err
		}

		ir := &InfoRaw{
			id:          container.ID,
			names:       container.Names,
			image:       container.Image,
			status:      inpected.State.Status,
			healthCheck: inpected.State.Health,
		}

		is := ir.ToString()
		if hc.only {
			if ir.healthCheck != nil {
				hc.infos = append(hc.infos, is)
			}
		} else {
			hc.infos = append(hc.infos, is)
		}
	}

	return hc.filter()
}

func (hc *HealthCheck) IsValid(healthCheck string) bool {

	for _, info := range hc.infos {
		if !info.HealthCheckIs(healthCheck) {
			return false
		}
	}

	return true
}

func (hc *HealthCheck) updateMeta() {
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
}

func (hc *HealthCheck) GetInfos() []*Info {
	return hc.infos
}

func (hc *HealthCheck) GetMeta() Meta {
	return hc.meta
}
