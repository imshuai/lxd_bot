package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/imshuai/sysutils"
	"gopkg.in/telebot.v3"
)

type tBot struct {
	db *bolt.DB
	*telebot.Bot
	cfg *config
}

var bot *tBot = &tBot{}

const (
	isDebug = true
)

func handleStart(c telebot.Context) error {
	msg := "欢迎使用本Bot！\n你可以发送 /create 命令来创建一个32MB内存的实例\n然后记得每天发送 /checkin 来签到"
	uname := c.Sender().Username
	uid := c.Sender().ID
	_, err := QueryUser(uid)
	if err == ErrorKeyNotFound {
		_, err = NewUser(uname, uid)
		if err != nil {
			return c.Send(err.Error())
		}
	} else if err != nil {
		return c.Send(err.Error())
	}
	return c.Send(msg)
}

func handleCreate(c telebot.Context) error {
	node := c.Args()[0]
	if nodes[node] == nil {
		return c.Reply("must specify correct node name")
	}
	u := c.Get("user").(*User)
	instance, err := u.CreateInstance(node, DefaultProfiles)
	if err != nil {
		return c.Reply(err.Error())
	}
	return c.Reply(fmt.Sprintf("创建成功!\n%s\n"+TplInstanceInformation+"\n%s", HR, instance.Name, instance.NodeName, instance.SSHPort, instance.NatPorts, instance.IPs(), HR))
}

func handleCheckin(c telebot.Context) error {
	var msg string
	u := c.Get("user").(*User)
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
	var instanceName string
	if i := c.Get("instanceName"); i == nil {
		instanceName = strings.Split(c.Args()[1], "@")[0]
		if instanceName == "" {
			return c.Send("must specify an instance name")
		}
	} else {
		instanceName = i.(string)
	}
	u := c.Get("user").(*User)
	err := u.Query()
	if err != nil {
		return err
	}

	if !u.HasInstance(instanceName) {
		return c.Send("This instance is not belong to you")
	}

	i := &Instance{Name: instanceName}

	err = i.Query()
	if err != nil {
		return c.Send("Query instance failed with error: " + err.Error() + "\nTry again later!")
	}
	msg := u.FormatInfo() + "\n" + HR

	instance, err := QueryInstance(i.NodeName, i.Name)
	if err != nil {
		return c.Reply("query instance failed with error: " + err.Error())
	}
	state, err := instance.State()
	if err != nil {
		return c.Send("Get instance state failed with error: " + err.Error() + "\nTry again later!")
	}
	msg = msg + "\n" + fmt.Sprintf("CPU: \n内存: %-5s\n磁盘: %-5s\n网络: 下行%-7s\t上行%-7s",
		sysutils.FormatSize(state.Memory.Usage),
		sysutils.FormatSize(state.Disk["root"].Usage),
		sysutils.FormatSize(state.Network["eth0"].Counters.BytesReceived),
		sysutils.FormatSize(state.Network["eth0"].Counters.BytesSent))
	markup := bot.NewMarkup()
	markup.Inline([]telebot.Row{
		{
			telebot.Btn{
				Text: "开机",
				Data: "!!control start " + string(i.Key()),
			},
			telebot.Btn{
				Text: "关机",
				Data: "!!control stop " + string(i.Key()),
			},
		},
		{
			telebot.Btn{
				Text: "重启",
				Data: "!!control restart " + string(i.Key()),
			},
			telebot.Btn{
				Text: "删机",
				Data: "!!control delete " + string(i.Key()),
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
	uid, _ := strconv.ParseInt(c.Args()[0], 10, 64)
	u, err := QueryUser(uid)
	if err != nil {
		return c.Reply("query user failed with error: " + err.Error())
	}
	u.IsManager = true
	err = u.Save()
	if err != nil {
		return c.Send(fmt.Sprintf("添加%s为管理员失败！\n%s\nError:%s", c.Sender().Username, HR, err.Error()))
	}
	return c.Send(fmt.Sprintf("添加%s为管理员成功！", c.Sender().Username))
}

func handleDeleteManager(c telebot.Context) error {
	uid, _ := strconv.ParseInt(c.Args()[0], 10, 64)
	u, err := QueryUser(uid)
	if err != nil {
		return c.Reply("query user failed with error: " + err.Error())
	}
	u.IsManager = false
	err = u.Save()
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

func handleAddNode(c telebot.Context) error {
	name := c.Args()[0]
	address := c.Args()[1]
	port := c.Args()[2]
	quota, _ := strconv.Atoi(c.Args()[3])
	node := &Node{
		Name:      name,
		Address:   address,
		Port:      port,
		LeftQuota: quota,
		MaxQuota:  quota,
		Instances: map[string]int64{},
		Users:     map[int64]string{},
		locker:    &sync.RWMutex{},
	}
	err := node.Save()
	if err != nil {
		return c.Send("Add node failed with error: " + err.Error())
	}
	conn, err := node.Connect(proxyClient)
	if err != nil {
		return c.Send("cannot connect to node with error: " + err.Error())
	}
	nodes[name] = conn
	return c.Send(fmt.Sprintf("创建新节点成功！\n节点名称：%s\n节点地址：%s\n节点端口：%s\n剩余配额：%d", name, address, port, quota))
}

func handleDeleteNode(c telebot.Context) error {
	name := c.Args()[0]
	node := &Node{Name: name}
	err := node.Query()
	if err != nil {
		return c.Reply("query node failed with error: " + err.Error())
	}
	err = node.Delete()
	if err != nil {
		return c.Reply("delete node failed with error: " + err.Error())
	}
	return nil
}
