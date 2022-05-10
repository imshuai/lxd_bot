package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/imshuai/sysutils"
)

const (
	USERS = "users"
)

var ExpirationTime = time.Hour * 24 * 7

type tUser struct {
	Name        string              `json:"name"`
	UID         int64               `json:"uid"`
	ChatID      int64               `json:"chat_id"`
	Created     sysutils.Time       `json:"created"`
	LastCheckin sysutils.Time       `json:"last_checkin"`
	Expiration  sysutils.Time       `json:"expiration"`
	SSHPort     int                 `json:"ssh_port"`
	UseableNum  int                 `json:"useable_num"`
	Instances   map[string]struct{} `json:"instances"`
	Profie      string              `json:"profie"`
	NodeName    string              `json:"node_name"`
	IsManager   bool                `json:"is_manager"`
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
	return fmt.Sprintf("用户ID: %d\n创建时间: %v\n签到时间: %v\n过期时间: %v\nSSH端口: %d\n剩余可用实例数量: %d",
		u.UID,
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
	t := sysutils.Now()
	u := &tUser{
		Name:        name,
		UID:         uid,
		ChatID:      chatID,
		Created:     t,
		LastCheckin: t,
		Expiration:  tExpiration(t),
		SSHPort:     0,
		UseableNum:  1,
		Instances:   make(map[string]struct{}),
		Profie:      "",
		NodeName:    name,
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
	return bot.db.View(func(tx *bolt.Tx) error {
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
		*u = *uu
		return nil
	})
}

func (u *tUser) Save() error {
	u.locker.Lock()
	defer u.locker.Unlock()
	return bot.db.Update(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists([]byte(USERS))
		if err != nil {
			return err
		}
		return bck.Put(u.Key(), []byte(u.String()))
	})
}

func (u *tUser) CreateInstance() error {
	return nil
	if u.UseableNum >= 1 {
		//TODO: 允许用户创建实例

	}
	return ErrorOverQuota
}

func (u *tUser) Checkin() error {
	u.locker.Lock()
	defer u.locker.Unlock()
	u.LastCheckin = sysutils.Now()
	u.Expiration = tExpiration(u.LastCheckin)
	err := bot.db.Update(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists([]byte(USERS))
		if err != nil {
			return err
		}
		return bck.Put(u.Key(), []byte(u.String()))
	})
	if err != nil {
		return ErrorCheckinFailed
	}
	return nil
}

func (u *tUser) HasInstance(name string) bool {
	_, ok := u.Instances[name]
	return ok
}
