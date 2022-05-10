package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/imshuai/sysutils"
	"gopkg.in/telebot.v3"
)

type tBot struct {
	db *bolt.DB
	*telebot.Bot
}

var bot *tBot = &tBot{}

const (
	isDebug = true
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
	inlineKeyboard := bot.NewMarkup()
	inlineKeyboard.Inline(telebot.Row{
		{
			Text: "管理实例",
			URL:  fmt.Sprintf("/control %s", u.Name),
		},
	})
	return c.Send(msg, inlineKeyboard)
}

func handleCheckin(c telebot.Context) error {
	var msg string
	u := c.Get("user").(*tUser)
	if len(u.Instances) <= 0 {
		msg = "你现在还没有可以续期的实例，无需签到"
	} else {
		err := u.Checkin()
		if err != nil {
			msg = "签到失败，发生错误\n" + HR + "\n" + err.Error() + "\n" + HR
		} else {
			msg = "签到成功\n" + HR + "\n" + u.FormatInfo() + "\n" + HR
		}
	}
	return c.Reply(msg)
}

func handlePing(c telebot.Context) error {
	msgTime := sysutils.Time{c.Message().Time().In(sysutils.SHANGHAI)}
	now := sysutils.Now()
	if msgTime.Add(time.Minute * 5).After(now.Time) {
		return c.Reply("我还活着! \n" + HR + "\n" + "消息时间: " + msgTime.String() + "\n响应时间: " + now.String() + "\n响应延迟: " + now.Sub(msgTime.Time).String())
	}
	return c.Reply("我活过来了！ \n" + HR + "\n" + "消息时间: " + msgTime.String() + "\n响应时间: " + now.String() + "\n响应延迟: " + now.Sub(msgTime.Time).String())
}

func handleInstanceControl(c telebot.Context) error {
	// TODO instance control
	if !c.Message().Private() {
		return c.Bot().Delete(c.Message())
	}
	u := &tUser{UID: c.Sender().ID}
	err := u.Get()
	if err != nil {
		return err
	}
	msg := u.FormatInfo() + "\n" + HR
	uuid := strings.Split(c.Args()[0], "@")[0]
	// if !u.HasInstance(uuid) {
	// 	return c.Send("该实例不属于你或实例UUID错误")
	// }
	state, err := GetInstanceState(uuid)
	if err != nil {
		return c.Send("获取实例状态失败，稍后再试")
	}
	profile, err := GetInstanceProfile(uuid)
	if err != nil {
		return c.Send("获取实例配置文件失败，稍后再试")
	}
	msg = msg + "\n" + fmt.Sprintf("CPU: \n内存: %-5s/%5s\n磁盘: %-5s/%5s\n网络: 下行%-7s\t上行%-7s",
		sysutils.FormatSize(state.Memory.Usage), profile.Config["limits.memory"],
		sysutils.FormatSize(state.Disk["root"].Usage), profile.Devices["root"]["size"],
		sysutils.FormatSize(state.Network["eth0"].Counters.BytesReceived),
		sysutils.FormatSize(state.Network["eth0"].Counters.BytesSent))
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

func handleAddManager(c telebot.Context) error {
	u := c.Get("user").(*tUser)
	u.IsManager = true
	err := u.Save()
	if err != nil {
		return c.Send(fmt.Sprintf("添加%s为管理员失败！\n%s\nError:%s", c.Sender().Username, HR, err.Error()))
	}
	return c.Send(fmt.Sprintf("添加%s为管理员成功！", c.Sender().Username))
}

func handleDeleteManager(c telebot.Context) error {
	u := c.Get("user").(*tUser)
	u.IsManager = false
	err := u.Save()
	if err != nil {
		return c.Send(fmt.Sprintf("删除%s管理员权限失败！\n%s\nError:%s", c.Sender().Username, HR, err.Error()))
	}
	return c.Send(fmt.Sprintf("删除%s管理员权限成功！", c.Sender().Username))
}
func handleGetUserList(c telebot.Context) error {

	return nil
}
func handleBanUser(c telebot.Context) error {
	return nil
}
func handleGetUserInfo(c telebot.Context) error {
	return nil
}
func handleDeleteInstance(c telebot.Context) error {
	return nil
}
