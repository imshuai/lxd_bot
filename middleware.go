package main

import (
	"fmt"

	"gopkg.in/telebot.v3"
)

func GetUserInfo(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		u := &tUser{UID: c.Sender().ID}
		err := u.Get()
		if err == ErrorKeyNotFound {
			return c.Send(fmt.Sprintf("未查询到用户信息, 请先给机器人@%s发送一条消息建立用户信息", bot.Me.Username))
		} else if err != nil {
			return c.Send("发生错误! \nError: " + err.Error())
		}
		c.Set("user", u)
		return next(c)
	}
}

func IsPrivateMessage(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		if !c.Message().Private() {
			msg, _ := bot.Send(c.Recipient(), fmt.Sprintf("该命令仅限于私聊bot使用"))
			c.Delete()
			return c.Bot().Delete(msg)
		}
		return next(c)
	}
}