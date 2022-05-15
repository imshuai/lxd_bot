package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type config struct {
	Token     string `json:"bot_token"`
	Proxy     string `json:"proxy"`
	DBPath    string `json:"db_path"`
	ChannelID string `json:"channel_id"`
	GroupID   string `json:"group_id"`
	CertFile  string `json:"cert_file"`
	KeyFile   string `json:"key_file"`
	AdminID   int64  `json:"admin_id"`
}

func readConfig(cfgPath string) *config {
	f, err := os.Open(cfgPath)
	switch err {
	case os.ErrNotExist:
		log.Fatalf("cannot read config from %s, file does not exist.\n", cfgPath)
	case os.ErrPermission:
		log.Fatalf("cannot read config from %s with permission issue.\n", cfgPath)
	}
	cfg := &config{}
	byts, _ := io.ReadAll(f)
	err = json.Unmarshal(byts, cfg)
	if err != nil {
		log.Fatalln("cannot unmarshal config with error:", err)
	}
	return cfg
}
