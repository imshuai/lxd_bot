package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/proxy"
	tele "gopkg.in/telebot.v3"
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
	botAPI, err := tele.NewBot(tele.Settings{
		Token:   cfg.Token,
		Poller:  &tele.LongPoller{Timeout: time.Second * 60},
		Client:  proxyClient,
		Verbose: isDebug,
	})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Authorized on account %s\n", botAPI.Me.Username)

	err = openDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("cannot open database with error: %s\n", err.Error())
	}
	go handleMessage(botAPI)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
	<-c
	bot.Quit()
}
