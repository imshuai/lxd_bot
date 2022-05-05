package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
	lxd "github.com/lxc/lxd/client"
	"golang.org/x/net/proxy"
	"gopkg.in/telebot.v3"
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

	var err error
	bot.Bot, err = telebot.NewBot(telebot.Settings{
		Token:   cfg.Token,
		Poller:  &telebot.LongPoller{Timeout: time.Second * 60},
		Client:  proxyClient,
		Verbose: isDebug,
	})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Authorized on account %s\n", bot.Me.Username)

	bot.db, err = bolt.Open(cfg.DBPath, 0664, bolt.DefaultOptions)
	if err != nil {
		log.Fatalf("cannot open database with error: %s\n", err.Error())
	}
	defer bot.db.Close()

	instance, err = lxd.ConnectLXD(cfg.ServerURL, &lxd.ConnectionArgs{
		TLSClientCert: func() string {
			byts, err := ioutil.ReadFile(cfg.CertFile)
			if err != nil {
				log.Fatalln(err.Error())
			}
			return string(byts)
		}(),
		TLSClientKey: func() string {
			byts, err := ioutil.ReadFile(cfg.KeyFile)
			if err != nil {
				log.Fatalln(err.Error())
			}
			return string(byts)
		}(),
		HTTPClient:         proxyClient,
		InsecureSkipVerify: true,
	})
	// instance, err = lxd.ConnectLXDUnix("", nil)
	if err != nil {
		//log.Fatalln("cannot connect to lxd server")
		log.Fatalln(err.Error())
	}
	defer instance.Disconnect()

	bot.Handle("/start", handleStart)
	bot.Handle("/create", handleCreate)
	bot.Handle("/checkin", handleCheckin)

	bot.Handle("/control", handleInstanceControl)

	bot.Handle("/ping", handlePing)

	bot.Handle(telebot.OnCallback, handleCallback)

	go bot.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
	<-c
	bot.Stop()
}
