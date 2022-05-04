package main

import (
	"encoding/json"

	"github.com/boltdb/bolt"
)

const (
	INSTANCE = "instances"
)

type tInstance struct {
	UUID        string `json:"uuid"`
	UserUUID    string `json:"user_uuid"`
	ProfileName string `json:"profile_name"`
	Created     tTime  `json:"created"`
	Expiration  tTime  `json:"expiration"`
	NodeName    string `json:"node_name"`
}

func (i *tInstance) Key() []byte {
	return []byte(i.UUID)
}

func (i *tInstance) String() string {
	byts, err := json.Marshal(i)
	if err != nil {
		return ""
	}
	return string(byts)
}

func (i *tInstance) Get() error {
	t := &tInstance{}
	return bot.View(func(tx *bolt.Tx) error {
		bck := tx.Bucket([]byte(INSTANCE))
		byts := bck.Get(i.Key())
		if byts == nil {
			return ErrorKeyNotFound
		}
		err := json.Unmarshal(byts, t)
		if err != nil {
			return ErrorUnmarshalInstance
		}
		i = t
		return nil
	})
}

func (i *tInstance) Save() error {
	return bot.Update(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists([]byte(INSTANCE))
		if err != nil {
			return ErrorBucketNameBlankOrTooLong
		}
		return bck.Put(i.Key(), []byte(i.String()))
	})
}
