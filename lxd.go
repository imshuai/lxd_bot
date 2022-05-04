package main

import (
	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
)

var instance lxd.InstanceServer

func restartInstance(name string) error {
	op, err := instance.UpdateInstanceState(name, api.InstanceStatePut{Action: "restart", Timeout: -1, Force: true}, "")
	if err != nil {
		return err
	}
	return op.Wait()
}

func startInstance(name string) error {
	op, err := instance.UpdateInstanceState(name, api.InstanceStatePut{Action: "start", Timeout: -1, Force: true}, "")
	if err != nil {
		return err
	}
	return op.Wait()
}

func stopInstance(name string) error {
	op, err := instance.UpdateInstanceState(name, api.InstanceStatePut{Action: "stop", Timeout: -1}, "")
	if err != nil {
		return err
	}
	return op.Wait()
}

func deleteInstance(name string) error {
	op, err := instance.DeleteInstance(name)
	if err != nil {
		return err
	}
	return op.Wait()
}
