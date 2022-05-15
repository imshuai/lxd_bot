package main

import (
	"math/rand"
	"time"

	"github.com/imshuai/sysutils"
)

const (
	HR = "------------------------"
)

func tExpiration(t sysutils.Time) sysutils.Time {
	return sysutils.Time{t.Add(ExpirationTime)}
}

type RandStringType int

const (
	RandStringTypeLetters RandStringType = iota
	RandStringTypeNumbers
	RandStringTypeLettersAndNumbers
	RandStringTypeAll
)

func RandString(size int, randType RandStringType) string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	symbols := "+-*.?!~#$%&()"
	str := ""
	switch randType {
	case RandStringTypeLetters:
		str = str + letters
	case RandStringTypeNumbers:
		str = str + numbers
	case RandStringTypeLettersAndNumbers:
		str = str + letters + numbers
	case RandStringTypeAll:
		str = str + letters + numbers + symbols
	}
	bytes := []byte(str)
	result := []byte{}
	rand.Seed(time.Now().UnixNano() + int64(rand.Intn(100)))
	for i := 0; i < size; i++ {
		result = append(result, bytes[rand.Intn(len(bytes))])
	}
	return string(result)
}
