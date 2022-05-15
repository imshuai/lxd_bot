package main

import (
	"flag"
	"fmt"
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
	"gopkg.in/telebot.v3/middleware"
)

var (
	nodes       map[string]lxd.InstanceServer = make(map[string]lxd.InstanceServer)
	proxyClient *http.Client
)

func main() {
	//读取配置文件
	cfgPath := flag.String("config", "config.json", "config file path")
	flag.Parse()
	bot.cfg = readConfig(*cfgPath)

	//构建代理连接
	switch strings.Split(bot.cfg.Proxy, ":")[0] {
	case "socks5":
		proxyDialer, err := proxy.SOCKS5("tcp", strings.TrimPrefix(bot.cfg.Proxy, "socks5://"), nil, proxy.Direct)
		if err != nil {
			log.Fatalf("[proxy]cannot parse proxy with error: %s\n", err)
		}
		proxyClient = &http.Client{Transport: &http.Transport{Dial: proxyDialer.Dial}}
	case "http", "https":
		proxyUrl, err := url.Parse(bot.cfg.Proxy)
		if err != nil {
			log.Fatalf("[proxy]cannot parse proxy with error: %s\n", err)
		}
		proxyClient = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	}

	var err error

	defer func() {
		for _, conn := range nodes {
			conn.Disconnect()
		}
	}()

	//创建telegram bot
	bot.Bot, err = telebot.NewBot(telebot.Settings{
		Token:   bot.cfg.Token,
		Poller:  &telebot.LongPoller{Timeout: time.Second * 60},
		Client:  proxyClient,
		Verbose: isDebug,
	})
	if err != nil {
		log.Fatalf("[bot]cannot create bot with error: %s\n", err.Error())
	}

	log.Printf("[bot]Authorized on account %s\n", bot.Me.Username)

	bot.db, err = bolt.Open(bot.cfg.DBPath, 0664, bolt.DefaultOptions)
	if err != nil {
		log.Fatalf("[database]cannot open database with error: %s\n", err.Error())
	}
	defer bot.db.Close()

	err = InitUser()
	if err != nil {
		log.Fatalln(err)
	}
	err = InitNode()
	if err != nil {
		log.Fatalln(err)
	}
	err = InitInstance()
	if err != nil {
		log.Fatalln(err)
	}

	bot.Handle("/start", handleStart, IsPrivateMessage)
	bot.Handle("/create", handleCreate, IsPrivateMessage, GetUserInfo)
	bot.Handle("/checkin", handleCheckin, GetUserInfo)
	bot.Handle("/control", handleInstanceControl, GetUserInfo, IsPrivateMessage)
	bot.Handle("/ping", handlePing)
	bot.Handle(telebot.OnCallback, handleCallback)

	manager := bot.Group()
	manager.Use(GetUserInfo, func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			u := c.Get("user").(*User)
			if u.IsManager || u.UID == bot.cfg.AdminID {
				return next(c)
			}
			msg, _ := bot.Reply(c.Message(), "请不要乱点管理员命令")
			warn, _ := bot.Send(c.Recipient(), fmt.Sprintf("/warn @%s", c.Sender().Username))
			c.Delete()
			bot.Delete(msg)
			return bot.Delete(warn)
		}
	})
	manager.Handle("/getuserlist", handleGetUserList, IsPrivateMessage)
	manager.Handle("/banuser", handleBanUser)
	manager.Handle("/getuserinfo", handleGetUserInfo)
	manager.Handle("/delinstance", handleDeleteInstance, IsPrivateMessage)

	administrator := bot.Group()
	administrator.Use(middleware.Whitelist(bot.cfg.AdminID))
	administrator.Handle("/addnode", handleAddNode)
	administrator.Handle("/delnode", handleDeleteNode)
	administrator.Handle("/addmanager", handleAddManager)
	administrator.Handle("/delmanager", handleDeleteManager)

	go bot.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
	<-c
	bot.Stop()
}
