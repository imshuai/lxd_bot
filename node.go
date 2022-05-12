package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	lxd "github.com/lxc/lxd/client"
)

type Node struct {
	Name      string              `json:"name"`
	Address   string              `json:"address"`
	Port      string              `json:"port"`
	LeftQuota int                 `json:"left_quota"`
	Instances map[string]struct{} `json:"instances"`
}

var (
	BckNodes = []byte("nodes")
)

func AddNode(n *Node) error {
	return bot.db.Update(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists(BckNodes)
		if err != nil {
			return err
		}
		byts, _ := json.Marshal(n)
		return bck.Put([]byte(n.Name), byts)
	})
}

func DeleteNode(name string) error {
	return bot.db.Update(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists(BckNodes)
		if err != nil {
			return err
		}
		node := &Node{}
		byts := bck.Get([]byte(name))
		if byts == nil {
			return ErrorKeyNotFound
		}
		err = json.Unmarshal(byts, node)
		if err != nil {
			return err
		}
		if len(node.Instances) > 0 {
			for k := range node.Instances {
				DeleteInstance(name, k)
			}
		}
		delete(nodes, name)
		return bck.Delete([]byte(name))
	})
}

func ConnectNode(addr, port string, client *http.Client) (lxd.InstanceServer, error) {
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
