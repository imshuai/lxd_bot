package main

import (
	"encoding/binary"
	"io/fs"
	"os"

	"github.com/boltdb/bolt"
)

var DB *bolt.DB

var (
	systemInfo = []byte("system-info")
	updateID   = []byte("update-id")
)

func openDB(dbPath string) error {
	d, err := bolt.Open(dbPath, fs.FileMode(os.O_RDWR|os.O_SYNC|os.O_CREATE), bolt.DefaultOptions)
	if err != nil {
		return err
	}
	DB = d
	return nil
}

func setUpdateID(id int) {
	DB.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists(systemInfo)
		bck := tx.Bucket(systemInfo)
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(id))
		return bck.Put(updateID, buf)
	})
}

func getUpdateID() int {
	return 0
}
