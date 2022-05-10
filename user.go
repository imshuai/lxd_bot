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

type User struct {
	Name        string              `json:"name"`
	UID         int64               `json:"uid"`
	Created     sysutils.Time       `json:"created"`
	LastCheckin sysutils.Time       `json:"last_checkin"`
	Expiration  sysutils.Time       `json:"expiration"`
	LeftQuota   int                 `json:"left_quota"`
	Instances   map[string]struct{} `json:"instances"`
	Profie      string              `json:"profie"`
	IsManager   bool                `json:"is_manager"`
	locker      *sync.RWMutex
}

func (u *User) String() string {
	byts, err := json.Marshal(u)
	if err != nil {
		return ""
	}
	return string(byts)
}

func (u *User) FormatInfo() string {
	return fmt.Sprintf("用户ID: %d\n创建时间: %v\n签到时间: %v\n过期时间: %v\n剩余可用实例数量: %d",
		u.UID,
		u.Created.String(),
		u.LastCheckin.String(),
		u.Expiration.String(),
		u.LeftQuota)
}

func ParseUser(str string) (*User, error) {
	u := &User{}
	err := json.Unmarshal([]byte(str), u)
	if err != nil {
		return nil, err
	}
	u.locker = &sync.RWMutex{}
	return u, nil
}

func NewUser(name string, uid int64, chatID int64) (*User, error) {
	t := sysutils.Now()
	u := &User{
		Name:        name,
		UID:         uid,
		Created:     t,
		LastCheckin: t,
		Expiration:  tExpiration(t),
		LeftQuota:   1,
		Instances:   map[string]struct{}{},
		Profie:      "",
		locker:      &sync.RWMutex{},
	}

	if err := u.Save(); err != nil {
		return nil, ErrorCreateUserFailed
	}
	return u, nil
}

func (u *User) Key() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(u.UID))
	return buf
}

func (u *User) Get() error {
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

func (u *User) Save() error {
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

func (u *User) CreateInstance() error {
	if u.LeftQuota >= 1 {
		//TODO: 允许用户创建实例
		return nil
	}
	return ErrorOverQuota
}

func (u *User) Checkin() error {
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

func (u *User) HasInstance(key string) bool {
	_, ok := u.Instances[key]
	return ok
}
