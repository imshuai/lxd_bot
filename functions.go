package main

import (
	"fmt"
	"time"
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

// 字节的单位转换 保留两位小数
func formatSize(fileSize int64) (size string) {
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
