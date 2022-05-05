package main

import (
	"time"

	"github.com/pborman/uuid"
)

var (
	SHANGHAI, _ = time.LoadLocation("Asia/Shanghai")
)

const (
	HR = "------------------------"
)

type tTime struct {
	time.Time
}

func (t tTime) MarshalText() ([]byte, error) {
	return []byte(t.Format("2006-01-02 15:04:05")), nil
}
func (t *tTime) UnmarshalText(str []byte) error {
	tt, err := time.Parse("2006-01-02 15:04:05", string(str))
	if err != nil {
		return err
	}
	*t = tTime{tt}
	return nil
}

func (t tTime) String() string {
	return t.Format("2006-01-02 15:04:05")
}

func tNow() tTime {
	return tTime{time.Now().Local().In(SHANGHAI)}
}

func tExpiration(t tTime) tTime {
	return tTime{t.Add(ExpirationTime)}
}

func getUUID() string {
	return uuid.New()
}
