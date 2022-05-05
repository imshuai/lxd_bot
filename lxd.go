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

func GetInstanceState(name string) (*api.InstanceState, error) {
	state, _, err := instance.GetInstanceState(name)
	return state, err
}

func GetInstanceProfile(name string) (*api.Profile, error) {
	i, _, err := instance.GetInstance(name)
	if err != nil {
		return nil, err
	}
	profile, _, err := instance.GetProfile(i.Profiles[0])
	if err != nil {
		return nil, err
	}
	return profile, nil
}
