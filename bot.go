package main

import (
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	"gopkg.in/telebot.v3"
)

type tBot struct {
	db *bolt.DB
	*telebot.Bot
}

var bot *tBot = &tBot{}

const (
	isDebug = false
)

func handleStart(c telebot.Context) error {
	msg := "欢迎使用本Bot！\n你可以发送 /create 命令来创建一个32MB内存的实例\n然后记得每天发送 /checkin 来签到"
	return c.Send(msg)
}

func handleCreate(c telebot.Context) error {
	// TODO 实例创建
	var msg string
	var err error
	uname := c.Sender().Username
	uid := c.Sender().ID
	chatid := c.Chat().ID
	u := &tUser{
		UID: uid,
	}
	err = u.Get()
	if err == ErrorKeyNotFound {
		u, err = NewUser(uname, uid, chatid)
		if err != nil {
			msg = err.Error()
			return c.Send(msg)
		}
	}
	err = u.CreateInstance()
	if err != nil {
		msg = err.Error()
		return c.Send(msg)
	}
	msg = "创建成功！" + HR + "\n" + u.FormatInfo() + "\n" + HR + fmt.Sprintf("\n务必于%s前签到", u.Expiration.String())
	return c.Send(msg)
}

func handleCheckin(c telebot.Context) error {
	var msg string
	u := &tUser{UID: c.Sender().ID}
	err := u.Get()
	if err == ErrorKeyNotFound {
		msg = "你现在还无法签到, 请先给机器人发送一条消息建立用户信息"
	} else {
		if len(u.Instances) <= 0 {
			msg = "你现在还没有可以续期的实例，无需签到"
		} else {
			err = u.Checkin()
			if err != nil {
				msg = "发生错误\n" + HR + "\n" + err.Error() + "\n" + HR
			} else {
				msg = "签到成功\n" + HR + "\n" + u.FormatInfo() + "\n" + HR
			}
		}
	}
	return c.Send(msg)
}

func handlePing(c telebot.Context) error {
	return c.Send("我还活着！")
}

func handleInstanceControl(c telebot.Context) error {
	// TODO instance control
	if !c.Chat().Private {
		return c.Bot().Delete(c.Message())
	}
	u := &tUser{UID: c.Sender().ID}
	err := u.Get()
	if err != nil {
		return err
	}
	uuid := strings.Split(c.Args()[0], "@")[0]
	if !u.HasInstance(uuid) {
		return c.Send("该实例不属于你或实例UUID错误")
	}
	state, err := GetInstanceState(uuid)
	if err != nil {
		return c.Send("获取实例状态失败，稍后再试")
	}
	profile, err := GetInstanceProfile(uuid)
	if err != nil {
		return c.Send("获取实例配置文件失败，稍后再试")
	}
	msg := fmt.Sprintf("CPU: \n内存: %-5s/%5s\n磁盘: %-5s/%5s\n网络: 下行%-7s\t上行%-7s",
		formatSize(state.Memory.Usage), profile.Config["limits.memory"],
		formatSize(state.Disk["root"].Usage), profile.Devices["root"]["size"],
		formatSize(state.Network["eth0"].Counters.BytesReceived),
		formatSize(state.Network["eth0"].Counters.BytesSent))
	markup := bot.NewMarkup()
	markup.Inline([]telebot.Row{
		{
			telebot.Btn{
				Unique: uuid,
				Text:   "开机",
				Data:   "start",
			},
			telebot.Btn{
				Unique: uuid,
				Text:   "关机",
				Data:   "stop",
			},
		},
		{
			telebot.Btn{
				Unique: uuid,
				Text:   "重启",
				Data:   "restart",
			},
			telebot.Btn{
				Unique: uuid,
				Text:   "删机",
				Data:   "delete",
			},
		},
	}...)
	return c.Send(msg, markup)

}

func handleCallback(c telebot.Context) error {
	// TODO inline keyboard callback
	return c.Edit(c.Callback().Data)
}
