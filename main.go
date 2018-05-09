package main

import (
	"github.com/cnrancher/docker-machine-driver-zstack/zstack"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(zstack.NewDriver("", ""))
}
