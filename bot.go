package main

import (
	"fmt"
	"log"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleNormalMessage(bot *tgapi.BotAPI, update *tgapi.Update) {
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	msg := tgapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	//msg.ReplyToMessageID = update.Message.MessageID

	bot.Send(msg)
}

func handleCommandMessage(bot *tgapi.BotAPI, update *tgapi.Update) {
	msg := tgapi.NewMessage(update.Message.Chat.ID, "")
	switch update.Message.CommandWithAt() {
	case "ping":
		msg.Text = "我还活着！"
	case "create":
		msg.Text = "现在还不能创建实例"
	case "checkin":
		msg.Text = "现在还不能签到"
	case "status":
		msg.Text = fmt.Sprintf("%s 你当前还没有实例\n@%s", update.Message.From.FirstName, update.Message.From.UserName)
	default:
		msg.Text = "欢迎使用本Bot！\n你可以发送 /create 命令来创建一个32MB内存的实例\n然后记得每天发送 /checkin 来签到"
	}
	bot.Send(msg)
}
