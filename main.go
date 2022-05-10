package main

import (
	"flag"
	"fmt"
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
	"github.com/urfave/cli"
	"golang.org/x/net/proxy"
	"gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

func main() {
	app := cli.NewApp()

	app.Run(os.Args)

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

	instance, err = lxd.ConnectLXD(cfg.Nodes[0].Address+":"+cfg.Nodes[0].Port, &lxd.ConnectionArgs{
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

	bot.Handle("/start", handleStart, IsPrivateMessage)
	bot.Handle("/create", handleCreate, IsPrivateMessage)
	bot.Handle("/checkin", handleCheckin, GetUserInfo)
	bot.Handle("/control", handleInstanceControl, GetUserInfo, IsPrivateMessage)
	bot.Handle("/ping", handlePing)

	manager := bot.Group()
	manager.Use(GetUserInfo, func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			u := c.Get("user").(*User)
			if u.IsManager {
				return next(c)
			}
			msg, _ := bot.Send(c.Recipient(), "请不要乱点管理员命令")
			warn, _ := bot.Send(c.Recipient(), fmt.Sprintf("/warn @%s", c.Sender().Username))
			c.Delete()
			bot.Delete(msg)
			return bot.Delete(warn)
		}
	})
	manager.Handle("/addmanager", handleAddManager, middleware.Whitelist(cfg.AdminID))
	manager.Handle("/delmanager", handleDeleteManager, middleware.Whitelist(cfg.AdminID))
	manager.Handle("/getuserlist", handleGetUserList, IsPrivateMessage)
	manager.Handle("/banuser", handleBanUser)
	manager.Handle("/getuserinfo", handleGetUserInfo)
	manager.Handle("/delinstance", handleDeleteInstance, IsPrivateMessage)

	bot.Handle(telebot.OnCallback, handleCallback)

	go bot.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
	<-c
	bot.Stop()
}
