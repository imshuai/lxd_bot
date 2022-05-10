package main

import (
	"encoding/json"
	"sync"

	"github.com/boltdb/bolt"
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

func (i *Instance) Update() error {
	return nil
}
