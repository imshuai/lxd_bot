package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pborman/uuid"
)

const (
	USERS = "users"
)

var ExpirationTime = time.Hour * 24 * 7

type tUser struct {
	Name        string   `json:"name"`
	UID         int64    `json:"uid"`
	UUID        string   `json:"uuid"`
	ChatID      int64    `json:"chat_id"`
	Created     tTime    `json:"created"`
	LastCheckin tTime    `json:"last_checkin"`
	Expiration  tTime    `json:"expiration"`
	SSHPort     int      `json:"ssh_port"`
	UseableNum  int      `json:"useable_num"`
	Instaces    []string `json:"instances"`
	locker      *sync.RWMutex
}

func (u *tUser) String() string {
	byts, err := json.Marshal(u)
	if err != nil {
		return ""
	}
	return string(byts)
}

func (u *tUser) FormatInfo() string {
	return fmt.Sprintf("用户ID: %d\nUUID: %s\n创建时间: %v\n签到时间: %v\n过期时间: %v\nSSH端口: %d\n剩余可用实例数量: %d",
		u.UID,
		u.UUID,
		u.Created.String(),
		u.LastCheckin.String(),
		u.Expiration.String(),
		u.SSHPort,
		u.UseableNum)
}

func ParseUser(str string) (*tUser, error) {
	u := &tUser{}
	err := json.Unmarshal([]byte(str), u)
	if err != nil {
		return nil, err
	}
	u.locker = &sync.RWMutex{}
	return u, nil
}

func NewUser(name string, uid int64, chatID int64) (*tUser, error) {
	t := tNow()
	u := &tUser{
		Name: name,
		UID:  uid,
		UUID: func() string {
			return uuid.New()
		}(),
		ChatID:      chatID,
		Created:     t,
		LastCheckin: t,
		Expiration:  tExpiration(t),
		SSHPort:     0,
		UseableNum:  1,
		Instaces:    []string{},
		locker:      &sync.RWMutex{},
	}

	if err := u.Save(); err != nil {
		return nil, ErrorCreateUserFailed
	}
	return u, nil
}

func (u *tUser) Key() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(u.UID))
	return buf
}

func (u *tUser) Get() error {
	err := bot.View(func(tx *bolt.Tx) error {
		bck := tx.Bucket([]byte(USERS))
		if bck == nil {
			return ErrorKeyNotFound
		}
		data := bck.Get(u.Key())
		if data == nil {
			return ErrorKeyNotFound
		}
		uu, err := ParseUser(string(data))
		if err != nil {
			return err
		}
		u = uu
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (u *tUser) Save() error {
	u.locker.Lock()
	defer u.locker.Unlock()
	return bot.Update(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists([]byte(USERS))
		if err != nil {
			return err
		}
		return bck.Put(u.Key(), []byte(u.String()))
	})
}

func (u *tUser) CreateInstance() (*tInstance, error) {
	if u.UseableNum >= 1 {
		//TODO: 允许用户创建实例
	}
	return nil, ErrorOverQuota
}

func (u *tUser) Checkin() error {
	u.locker.Lock()
	defer u.locker.Unlock()
	u.LastCheckin = tNow()
	u.Expiration = tExpiration(u.LastCheckin)
	err := bot.Update(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists([]byte(USERS))
		if err != nil {
			return err
		}
		return bck.Put(u.Key(), []byte(u.String()))
	})
	if err != nil {
		return ErrorCheckinFailed
	}
	for _, instanceName := range u.Instaces {
		i := &tInstance{UUID: instanceName}
		err := i.Get()
		if err != nil {
			log.Printf("[%12s][Checkin][%7s] %s", u.Name, "ERROR", err.Error())
			continue
		}
		i.Expiration = u.Expiration
		err = i.Save()
		if err != nil {
			log.Printf("[%12s][Checkin][%7s] %s", u.Name, "ERROR", err.Error())
		}
	}
	return nil
}
