package main

import (
	"encoding/json"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/lxc/lxd/shared/api"
)

var (
	BckInstances = []byte("instances")
)

type Instance struct {
	Name     string              `json:"name"`
	Profiles map[string]struct{} `json:"profiles"`
	SSHPort  int                 `json:"ssh_port"`
	NatPorts string              `json:"nat_ports"`
	IPv4     []string            `json:"ipv4"`
	IPv6     []string            `json:"ipv6"`
	NodeName string              `josn:"node_name"`
	Username string              `json:"username"`
	locker   *sync.RWMutex
}

func (i *Instance) Key() []byte {
	i.locker.RLock()
	defer i.locker.RUnlock()
	return []byte(i.NodeName + ":" + i.Name)
}

func (i *Instance) Save() error {
	i.locker.Lock()
	defer i.locker.Unlock()
	return bot.db.Update(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists(BckInstances)
		if err != nil {
			return err
		}
		byts, err := json.Marshal(i)
		if err != nil {
			return err
		}
		return bck.Put(i.Key(), byts)
	})
}

func (i *Instance) Query() error {
	i.locker.Lock()
	defer i.locker.Unlock()
	return bot.db.View(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists(BckInstances)
		if err != nil {
			return err
		}
		byts := bck.Get([]byte(i.Key()))
		return json.Unmarshal(byts, i)
	})
}

func (i *Instance) Update() error {
	return nil
}

func DeleteInstance(node, name string) error {
	op, err := nodes[node].DeleteInstance(name)
	if err != nil {
		return err
	}
	err = op.Wait()
	if err != nil {
		return err
	}
	// TODO delete instance info in database
	return nil
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
