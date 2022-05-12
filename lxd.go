package main

import (
	"io/ioutil"
	"log"
	"net/http"

	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
)

func RestartInstance(node, name string) error {
	op, err := nodes[node].UpdateInstanceState(name, api.InstanceStatePut{Action: "restart", Timeout: -1, Force: true}, "")
	if err != nil {
		return err
	}
	return op.Wait()
}

func StartInstance(node, name string) error {
	op, err := nodes[node].UpdateInstanceState(name, api.InstanceStatePut{Action: "start", Timeout: -1, Force: true}, "")
	if err != nil {
		return err
	}
	return op.Wait()
}

func StopInstance(node, name string) error {
	op, err := nodes[node].UpdateInstanceState(name, api.InstanceStatePut{Action: "stop", Timeout: -1}, "")
	if err != nil {
		return err
	}
	return op.Wait()
}

func DeleteInstance(node, name string) error {
	op, err := nodes[node].DeleteInstance(name)
	if err != nil {
		return err
	}
	return op.Wait()
}

func CreateInstance(node, name, profile string) (err error) {
	op, err := nodes[node].CreateInstance(api.InstancesPost{
		InstancePut: api.InstancePut{
			Architecture: "",
			Config:       map[string]string{},
			Devices:      map[string]map[string]string{},
			Ephemeral:    false,
			Profiles:     []string{profile},
			Restore:      "",
			Stateful:     false,
			Description:  "",
		},
		Name:         name,
		Source:       api.InstanceSource{Type: "image", Alias: "alpine-base"},
		InstanceType: "",
		Type:         api.InstanceTypeContainer,
	})
	if err != nil {
		return err
	}
	return op.Wait()
}

func GetInstanceState(node, name string) (*api.InstanceState, error) {
	state, _, err := nodes[node].GetInstanceState(name)
	return state, err
}

func GetInstanceProfile(node, name string) (*api.Profile, error) {
	i, _, err := nodes[node].GetInstance(name)
	if err != nil {
		return nil, err
	}
	profile, _, err := nodes[node].GetProfile(i.Profiles[0])
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func ConnectLXD(addr, port string, client *http.Client) (lxd.InstanceServer, error) {
	return lxd.ConnectLXD(addr+":"+port, &lxd.ConnectionArgs{
		TLSClientCert: func() string {
			byts, err := ioutil.ReadFile(bot.cfg.CertFile)
			if err != nil {
				log.Fatalln(err.Error())
			}
			return string(byts)
		}(),
		TLSClientKey: func() string {
			byts, err := ioutil.ReadFile(bot.cfg.KeyFile)
			if err != nil {
				log.Fatalln(err.Error())
			}
			return string(byts)
		}(),
		HTTPClient:         client,
		InsecureSkipVerify: true,
	})
}
