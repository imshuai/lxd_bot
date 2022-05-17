package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	lxd "github.com/lxc/lxd/client"
)

type Node struct {
	Name      string           `json:"name"`
	Address   string           `json:"address"`
	Port      string           `json:"port"`
	MaxQuota  int              `json:"max_quota"`
	LeftQuota int              `json:"left_quota"`
	Instances map[string]int64 `json:"instances"`
	Users     map[int64]string `json:"users"`
}

var (
	BckNodes = []byte("nodes")
)

func InitNode() error {
	return bot.db.Update(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists(BckNodes)
		if err != nil {
			return err
		}
		return bck.ForEach(func(k, v []byte) error {
			node := &Node{}
			err := json.Unmarshal(v, node)
			if err != nil {
				return err
			}
			conn, err := node.Connect(proxyClient)
			if err != nil {
				return err
			}
			nodes[node.Name] = conn
			return nil
		})
	})
}

func (n *Node) Key() []byte {
	return []byte("node:" + n.Name)
}

func (n *Node) Query() error {
	return bot.db.View(func(tx *bolt.Tx) error {
		bck := tx.Bucket(BckNodes)
		byts := bck.Get([]byte(n.Key()))
		return json.Unmarshal(byts, n)
	})
}

func (n *Node) Save() error {
	return bot.db.Update(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists(BckNodes)
		if err != nil {
			return err
		}
		byts, err := json.Marshal(n)
		if err != nil {
			return err
		}
		return bck.Put(n.Key(), byts)
	})
}

func (n *Node) Delete() error {
	if n.Name == "" || n.Address == "" || n.Port == "" {
		return errors.New("must specify name, address and port")
	}
	err := n.Query()
	if err != nil {
		return err
	}
	for uid, instanceName := range n.Users {
		u := &User{UID: uid}
		err = u.Query()
		if err != nil {
			return err
		}
		err = u.DeleteInstance(instanceName, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) Connect(client *http.Client) (lxd.InstanceServer, error) {
	return lxd.ConnectLXD("https://"+n.Address+":"+n.Port, &lxd.ConnectionArgs{
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

func (n *Node) IsExist() bool {
	err := bot.db.View(func(tx *bolt.Tx) error {
		bck := tx.Bucket(BckNodes)
		byts := bck.Get(n.Key())
		if byts == nil {
			return ErrorKeyNotFound
		}
		return nil
	})
	return err != ErrorKeyNotFound
}

func QueryNode(name string) (*Node, error) {
	node := &Node{
		Name:      name,
		Address:   "",
		Port:      "",
		MaxQuota:  0,
		LeftQuota: 0,
		Instances: map[string]int64{},
		Users:     map[int64]string{},
	}
	err := node.Query()
	if err != nil {
		return nil, err
	}
	return node, nil
}

func NewNode(name, address, port string, quota int) (*Node, error) {
	node := &Node{
		Name:      name,
		Address:   address,
		Port:      port,
		MaxQuota:  quota,
		LeftQuota: quota,
		Instances: map[string]int64{},
		Users:     map[int64]string{},
	}
	if node.IsExist() {
		return nil, errors.New("node name already exist")
	}
	err := node.Save()
	if err != nil {
		return nil, err
	}
	return node, nil
}
