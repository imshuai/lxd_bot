package main

import (
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

var (
	SHANGHAI, _ = time.LoadLocation("Asia/Shanghai")
)

const (
	HR = "------------------------"
)

type tTime time.Time

func (t tTime) MarshalText() ([]byte, error) {
	return []byte(time.Time(t).Format("2006-01-02 15:04:05")), nil
}
func (t tTime) UnmarshalText(str []byte) error {
	tt, err := time.Parse("2006-01-02 15:04:05", string(str))
	if err != nil {
		return err
	}
	t = tTime(tt)
	return nil
}

func (t tTime) String() string {
	return time.Time(t).Format("2006-01-02 15:04:05")
}

type botDB struct {
	*bolt.DB
	updateID int
	cfg      *config
	locker   *sync.RWMutex
}

var bot = &botDB{
	locker: &sync.RWMutex{},
}

func tNow() tTime {
	return tTime(time.Now().Local().In(SHANGHAI))
}

func tExpiration(t tTime) tTime {
	return tTime(time.Time(t).Add(ExpirationTime))
}
