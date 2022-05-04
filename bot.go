package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/boltdb/bolt"
	TGAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	systemInfo = []byte("system-info")
	updateID   = []byte("update-id")
)

const (
	isDebug = false
)

func openDB(dbPath string) error {
	d, err := bolt.Open(dbPath, 0664, bolt.DefaultOptions)
	if err != nil {
		return err
	}
	bot.DB = d
	return nil
}

func (bot *botDB) SetUpdateID(id int) error {
	bot.locker.Lock()
	defer bot.locker.Unlock()
	if id <= bot.updateID {
		return errors.New("update ID is smaller than current")
	}
	err := bot.Update(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists(systemInfo)
		if err != nil {
			return err
		}
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(id))
		return bck.Put(updateID, buf)
	})
	if err != nil {
		return err
	}
	bot.updateID = id
	return nil
}

func (bot *botDB) GetUpdateID() int {
	bot.locker.Lock()
	defer bot.locker.Unlock()
	if bot.updateID != 0 {
		return bot.updateID
	}
	var id int
	err := bot.View(func(tx *bolt.Tx) error {
		bck, err := tx.CreateBucketIfNotExists(systemInfo)
		if err != nil {
			return err
		}
		buf := bck.Get(updateID)
		id = int(binary.LittleEndian.Uint32(buf))
		return nil
	})
	if err != nil {
		return 0
	}
	bot.updateID = id
	return bot.updateID
}

func handleNormalMessage(botAPI *TGAPI.BotAPI, update *TGAPI.Update) {
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	msg := TGAPI.NewMessage(update.Message.Chat.ID, update.Message.Text)
	//msg.ReplyToMessageID = update.Message.MessageID

	botAPI.Send(msg)
}

func handleCommandMessage(botAPI *TGAPI.BotAPI, update *TGAPI.Update) {
	msg := TGAPI.NewMessage(update.Message.Chat.ID, "")
	switch strings.Trim(strings.Split(update.Message.CommandWithAt(), "@")[0], " ") {
	case "ping":
		msg.Text = "我还活着！" + "\n" + update.Message.CommandArguments()
	case "create":
		msg.ReplyMarkup = TGAPI.ReplyKeyboardRemove{RemoveKeyboard: true}
		msg.Text = "test button"
	case "checkin":
		u := &tUser{UID: update.Message.From.ID}
		err := u.Get()
		if err == ErrorKeyNotFound {
			msg.Text = "你现在还无法签到, 请先给机器人发送一条消息建立用户信息"
		} else {
			if len(u.Instaces) <= 0 {
				msg.Text = "你现在还没有可以续期的实例，无需签到"
			} else {
				err = u.Checkin()
				if err != nil {
					msg.Text = "发生错误\n" + HR + "\n" + err.Error() + "\n" + HR
				} else {
					msg.Text = "签到成功\n" + HR + "\n" + u.FormatInfo() + "\n" + HR
				}
			}
		}
	case "status":
		msg.Text = fmt.Sprintf("%s 你当前还没有实例\n@%s", update.Message.From.FirstName, update.Message.From.UserName)
	case "start":
		msg.Text = "欢迎使用本Bot！\n你可以发送 /create 命令来创建一个32MB内存的实例\n然后记得每天发送 /checkin 来签到"
		uname := update.Message.From.UserName
		uid := update.Message.From.ID
		chatid := update.Message.Chat.ID
		t := &tUser{
			UID: uid,
		}
		err := t.Get()
		if err == ErrorKeyNotFound {
			u, err := NewUser(uname, uid, chatid)
			if err != nil {
				msg.Text = msg.Text + "\n\n" + err.Error()
			} else {
				msg.Text = msg.Text + "\n\n" + u.FormatInfo()
			}
		} else {
			msg.Text = msg.Text + "\n\n" + t.FormatInfo()
		}
	}
	botAPI.Send(msg)
}

func handleMessage(botAPI *TGAPI.BotAPI) {
	u := TGAPI.NewUpdate(bot.GetUpdateID())
	u.Timeout = 60

	updates := botAPI.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {

			if update.Message.IsCommand() {
				go handleCommandMessage(botAPI, &update)
				continue
			}
			go handleNormalMessage(botAPI, &update)
		}
	}
}

//Quit handle bot shutdown
func (bot *botDB) Quit() {
	bot.SetUpdateID(bot.updateID)
	bot.DB.Close()
}
