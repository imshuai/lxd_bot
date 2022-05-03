package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"strings"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/proxy"
)

func main() {
	cfgPath := flag.String("config", "config.json", "config file path")
	flag.Parse()
	cfg := readConfig(*cfgPath)
	var proxyClient *http.Client
	switch strings.Split(cfg.Proxy, ":")[0] {
	case "socks5":
		proxyDialer, err := proxy.SOCKS5("tcp", strings.TrimPrefix(cfg.Proxy, "socks5://"), nil, proxy.Direct)
		if err != nil {
			log.Fatalf("cannot parse proxy with error: %s\n", err)
		}
		proxyClient = &http.Client{Transport: &http.Transport{Dial: proxyDialer.Dial}}
	case "http", "https":
		proxyUrl, err := url.Parse(cfg.Proxy)
		if err != nil {
			log.Fatalf("cannot parse proxy with error: %s\n", err)
		}
		proxyClient = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	}
	bot, err := tgapi.NewBotAPIWithClient(cfg.Token, tgapi.APIEndpoint, proxyClient)
	if err != nil {
		log.Fatalln(err)
	}
	bot.Debug = true

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	err = openDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("cannot open database with error: %s\n", err.Error())
	}

	u := tgapi.NewUpdate(getUpdateID())
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {

			if update.Message.IsCommand() {
				go handleCommandMessage(bot, &update)
				continue
			}
			go handleNormalMessage(bot, &update)
		}
	}
}
