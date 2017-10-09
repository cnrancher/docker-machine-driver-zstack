package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/rancher/docker-machine-driver-zstack/zstack"
)

func main() {
	plugin.RegisterDriver(zstack.NewDriver())
}
