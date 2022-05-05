package main

import (
	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
)

var instance lxd.InstanceServer

func RestartInstance(name string) error {
	op, err := instance.UpdateInstanceState(name, api.InstanceStatePut{Action: "restart", Timeout: -1, Force: true}, "")
	if err != nil {
		return err
	}
	return op.Wait()
}

func StartInstance(name string) error {
	op, err := instance.UpdateInstanceState(name, api.InstanceStatePut{Action: "start", Timeout: -1, Force: true}, "")
	if err != nil {
		return err
	}
	return op.Wait()
}

func StopInstance(name string) error {
	op, err := instance.UpdateInstanceState(name, api.InstanceStatePut{Action: "stop", Timeout: -1}, "")
	if err != nil {
		return err
	}
	return op.Wait()
}

func DeleteInstance(name string) error {
	op, err := instance.DeleteInstance(name)
	if err != nil {
		return err
	}
	return op.Wait()
}

func CreateInstance(profile string) (uuid string, err error) {
	op, err := instance.CreateInstance(api.InstancesPost{
		Name: getUUID(),
		Source: api.InstanceSource{
			Type:  "image",
			Alias: "base",
		},
		Type: api.InstanceTypeContainer,
	})
	if err != nil {
		return "", err
	}
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
