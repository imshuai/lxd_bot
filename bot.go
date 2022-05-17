package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"gopkg.in/telebot.v3"
)

type tBot struct {
	db *bolt.DB
	*telebot.Bot
	cfg *config
}

var bot *tBot = &tBot{}

const (
	isDebug = false
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
	if len(c.Args()) < 1 {
		nodelist := ""
		for k := range nodes {
			node, err := QueryNode(k)
			if err != nil {
				c.Reply(err.Error())
			}
			nodelist = nodelist + fmt.Sprintf("%-5s: %3d\n", node.Name, node.LeftQuota)
		}
		usage := "需要指定要创建实例的节点名称，例如 /create hk1\n当前所有节点剩余配额如下: \n" + HR + "\n"
		return c.Reply(usage + nodelist)
	}
	nodeName := c.Args()[0]
	if nodes[nodeName] == nil {
		return c.Reply("must specify correct node name")
	}
	u := c.Get("user").(*User)
	instance, err := u.CreateInstance(nodeName, DefaultProfiles)
	if err != nil {
		return c.Send(err.Error())
	}
	node, _ := QueryNode(instance.NodeName)
	return c.Send(fmt.Sprintf("创建成功!\n%s\n"+TplInstanceInformation+"\n公网地址: %s\n%s", HR, instance.Name, instance.NodeName, instance.SSHPort, instance.NatPorts, instance.IPs(), node.Address, HR))
}

func handlePing(c telebot.Context) error {
	msgTime := Time{c.Message().Time().In(SHANGHAI)}
	now := Now()
	if msgTime.Add(time.Minute * 5).After(now.Time) {
		return c.Reply("我还活着! \n" + HR + "\n" + "消息时间: " + msgTime.String() + "\n响应时间: " + now.String() + "\n响应延迟: " + now.Sub(msgTime.Time).String())
	}
	return c.Reply("我活过来了！ \n" + HR + "\n" + "消息时间: " + msgTime.String() + "\n响应时间: " + now.String() + "\n响应延迟: " + now.Sub(msgTime.Time).String())
}

func handleListInstance(c telebot.Context) error {
	u := c.Get("user").(*User)
	err := u.Query()
	if err != nil {
		return c.Reply(err.Error())
	}
	msg := ""
	for instanceName, nodeName := range u.Instances {
		instance, err := QueryInstance(nodeName, instanceName)
		if err != nil {
			return c.Reply(err.Error())
		}
		state, err := instance.State()
		if err != nil {
			return c.Reply(err.Error())
		}
		msg = msg + fmt.Sprintf("%-12s: %6s\n", instanceName, state.Status)
	}
	return c.Reply(fmt.Sprintf("%-20s\n下面是你的实例列表：\n%s", c.Sender().FirstName, msg))
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
	users, err := QueryUsers()
	if err != nil {
		return c.Reply("query user failed with error: " + err.Error())
	}
	userlist := ""
	for _, user := range users {
		instances := ""
		for instanceName := range user.Instances {
			instances += instanceName + ", "
		}
		instances = strings.TrimRight(instances, ", ")
		userlist += fmt.Sprintf("@%-15s: %s\n", user.Name, instances)
	}
	userlist = strings.TrimRight(userlist, "\n")
	return c.Reply(fmt.Sprintf("所有用户及其实例如下：\n%s", userlist))
}

func handleDeleteInstance(c telebot.Context) error {
	if len(c.Args()) < 1 {
		return c.Reply("需要指定要删除实例的名称，例如 /delinstance hk1-88888")
	}
	instanceName := c.Args()[0]
	nodeName := strings.Split(instanceName, "-")[0]
	instance, err := QueryInstance(nodeName, instanceName)
	if err != nil {
		return c.Reply("query instance failed with error: " + err.Error())
	}
	err = instance.Delete(false)
	if err != nil {
		return c.Reply("delete instance failed with error: " + err.Error())
	}
	return c.Reply("delete instance success")
}

func handleAddNode(c telebot.Context) error {
	name := c.Args()[0]
	address := c.Args()[1]
	if strings.HasPrefix(address, "http") {
		address = strings.TrimPrefix(address, "http://")
		address = strings.TrimPrefix(address, "https://")
	}
	port := c.Args()[2]
	quota, _ := strconv.Atoi(c.Args()[3])
	node, err := NewNode(name, address, port, quota)
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
