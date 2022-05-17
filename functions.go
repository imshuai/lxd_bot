package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	HR = "------------------------"
)

var (
	SHANGHAI, _ = time.LoadLocation("Asia/Shanghai")
)

func tExpiration(t Time) Time {
	return Time{t.Add(ExpirationTime)}
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

type Time struct {
	time.Time
}

func (t Time) MarshalText() ([]byte, error) {
	return []byte(t.Format("2006-01-02 15:04:05")), nil
}

func (t *Time) UnmarshalText(str []byte) error {
	tt, err := time.Parse("2006-01-02 15:04:05", string(str))
	if err != nil {
		return err
	}
	*t = Time{tt}
	return nil
}

func (t Time) String() string {
	return t.Format("2006-01-02 15:04:05")
}

func Now() Time {
	return Time{time.Now().Local().In(SHANGHAI)}
}

// 字节的单位转换 保留两位小数
func FormatSize(fileSize int64) (size string) {
	if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}
