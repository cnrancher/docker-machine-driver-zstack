package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/orangedeng/docker-machine-driver-zstack/zstack"
)

func main() {
	plugin.RegisterDriver(zstack.NewDriver("",""))
}
