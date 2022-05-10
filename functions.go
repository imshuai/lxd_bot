package main

import (
	"github.com/imshuai/sysutils"
)

const (
	HR = "------------------------"
)

func tExpiration(t sysutils.Time) sysutils.Time {
	return sysutils.Time{t.Add(ExpirationTime)}
}
