package healthcheck

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
)

type Info struct {
	id          string
	names       string
	image       string
	status      string
	healthCheck string
}

func (i *Info) HealthCheckIs(need string) bool {
	return i.healthCheck == need
}

func (i *Info) GetHealthCheck() string {
	return i.healthCheck
}

func (i *Info) Print(hcm Meta) string {
	hcms := hcm.ToString()
	return fmt.Sprintf("%-"+hcms.id+"s"+
		"%-"+hcms.names+"s"+
		"%-"+hcms.image+"s"+
		"%-"+hcms.status+"s"+
		"%-"+hcms.healthCheck+"s",
		i.id, i.names, i.image, i.status, i.healthCheck,
	)
}

type InfoRaw struct {
	id          string
	names       []string
	image       string
	status      string
	healthCheck *types.Health
}

func (ir *InfoRaw) ToString() *Info {
	return &Info{
		id:          ir.getID(),
		names:       ir.getCanonicalContainerName(),
		image:       ir.getImageNameOrHash(),
		status:      ir.status,
		healthCheck: ir.translateHealthCheck(),
	}
}

func (ir *InfoRaw) getImageNameOrHash() string {
	if strings.HasPrefix(ir.image, "sha256:") {
		return ir.image[7:19]
	}

	return ir.image
}

func (ir *InfoRaw) getID() string {
	return ir.id[0:12]
}

func (ir *InfoRaw) getCanonicalContainerName() string {
	for _, name := range ir.names {
		if strings.LastIndex(name, "/") == 0 {
			return name[1:]
		}
	}

	return ir.names[0][1:]
}

func (ir *InfoRaw) translateHealthCheck() string {
	if ir.healthCheck != nil {
		return ir.healthCheck.Status
	}

	return "-"
}
