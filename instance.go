package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/lxc/lxd/shared/api"
)

var (
	BckInstances = []byte("instances")
)

type Instance struct {
	Name     string   `json:"name"`
	SSHPort  string   `json:"ssh_port"`
	NatPorts string   `json:"nat_ports"`
	IPv4     string   `json:"ipv4"`
	IPv6     string   `json:"ipv6"`
	Profiles []string `json:"profiles"`
	NodeName string   `josn:"node_name"`
	UserID   int64    `json:"user_id"`
	locker   *sync.RWMutex
}

func InitInstance() error {
	return bot.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(BckInstances)
		return err
	})
}

//Key generate this instance's key
func (i *Instance) Key() []byte {
	i.locker.RLock()
	defer i.locker.RUnlock()
	return []byte(i.NodeName + ":" + i.Name)
}

//Save store this instance to the database
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

//Query get this instance's information from database
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

//Delete this instance from lxd server and database
func (i *Instance) Delete(ignoreLXDError bool) error {
	if i.Name == "" || i.NodeName == "" {
		return errors.New("must specify instance name and node name")
	}
	op, err := nodes[i.NodeName].DeleteInstance(i.Name)
	if err != nil || !ignoreLXDError {
		return err
	}
	err = op.Wait()
	if err != nil || !ignoreLXDError {
		return err
	}
	err = i.Query()
	if err != nil {
		return err
	}
	return bot.db.Update(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists(BckInstances)
		if err != nil {
			return err
		}
		return bck.Delete(i.Key())
	})
}

func (i *Instance) Create() error {
	i.locker.Lock()
	defer i.locker.Unlock()
	conn := nodes[i.NodeName]
	if conn == nil {
		return errors.New("cannot connect to node: " + i.NodeName)
	}
	node, err := QueryNode(i.NodeName)
	if err != nil {
		return err
	}
	node.locker.Lock()
	defer node.locker.Unlock()
	i.IPv4 = fmt.Sprintf("10.10.11.1%0d", node.MaxQuota-node.LeftQuota+1)
	i.SSHPort = strings.Split(i.IPv4, ".")[3] + "00"
	i.NatPorts = strings.Split(i.IPv4, ".")[3] + "01-" + strings.Split(i.IPv4, ".")[3] + "20"
	op, err := conn.CreateInstance(api.InstancesPost{
		InstancePut: api.InstancePut{
			Architecture: "",
			Config:       map[string]string{},
			Devices: map[string]map[string]string{
				"eth0": {
					"type":         "nic",
					"limits.max":   "50Mbit",
					"network":      "lxdbr0",
					"ipv4.address": i.IPv4,
				},
				"ssh-port": {
					"bind":    "host",
					"connect": "tcp:127.0.0.1:22",
					"listen":  "tcp:0.0.0.0:" + i.SSHPort,
					"type":    "proxy",
				},
				"nat-tcp-ports": {
					"bind":    "host",
					"connect": "tcp:127.0.0.1:" + i.NatPorts,
					"listen":  "tcp:0.0.0.0:" + i.NatPorts,
					"type":    "proxy",
				},
				"nat-udp-ports": {
					"bind":    "host",
					"connect": "tcp:127.0.0.1:" + i.NatPorts,
					"listen":  "tcp:0.0.0.0:" + i.NatPorts,
					"type":    "proxy",
				},
			},
			Ephemeral:   false,
			Profiles:    i.Profiles,
			Restore:     "",
			Stateful:    false,
			Description: "",
		},
		Name:         i.Name,
		Source:       api.InstanceSource{Type: "image", Alias: "alpine-base"},
		InstanceType: "",
		Type:         api.InstanceTypeContainer,
	})
	if err != nil {
		return err
	}
	err = op.Wait()
	if err != nil {
		return err
	}
	node.LeftQuota -= 1
	node.Instances[i.Name] = i.UserID
	node.Users[i.UserID] = i.Name
	err = node.Save()
	//TODO handle node and instance information save error
	if err != nil {
		return err
	}
	err = i.Save()
	if err != nil {
		return err
	}
	return nil
}

func (i *Instance) State() (*api.InstanceState, error) {
	node := nodes[i.NodeName]
	if node == nil {
		return nil, errors.New("must specify correct node name or node cannot visit now")
	}
	state, _, err := node.GetInstanceState(i.Name)
	return state, err
}

func (i *Instance) IPs() string {
	return i.IPv4 + "\n" + i.IPv6
}

func QueryInstance(node, name string) (*Instance, error) {
	instance := &Instance{Name: name, NodeName: node}
	err := instance.Query()
	if err != nil {
		return nil, err
	}
	return instance, nil
}
