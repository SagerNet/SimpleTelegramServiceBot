package main

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

import (
	"encoding/json"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/v2fly/v2ray-core/v4/common"
	"io/ioutil"
	"time"
)

func main() {
	configBytes, err := ioutil.ReadFile("config.json")
	common.Must(err)
	config := &Config{}
	common.Must(json.Unmarshal(configBytes, config))
	bot, err := tg.NewBotAPI(config.BotToken)
	common.Must(err)
	service := &Service{config, bot}
	service.loopUpdates()
}

type Config struct {
	BotToken string `json:"botToken"`
}

type Service struct {
	*Config
	*tg.BotAPI
}

func (bot *Service) loopUpdates() {
	updates := bot.GetUpdatesChan(tg.UpdateConfig{AllowedUpdates: []string{"message"}})
	for update := range updates {
		go bot.onUpdate(update)
	}
}

func (bot *Service) onUpdate(update tg.Update) {
	if update.Message != nil {
		bot.onNewMessage(update.Message)
	}
}

func (bot *Service) onNewMessage(message *tg.Message) {
	if !message.Chat.IsSuperGroup() {
		return
	}
	{
		// channel message check
		if message.SenderChat != nil {
			chat, err := bot.GetChat(tg.ChatInfoConfig{
				ChatConfig: tg.ChatConfig{
					ChatID: message.Chat.ID,
				},
			})
			if err != nil {
				newError("failed to get chat").Base(err).WriteToLog()
			}

			if message.SenderChat.ID == message.Chat.ID || message.SenderChat.ID == chat.LinkedChatID {
				// admin
				return
			}

			err = bot.MustRequests(
				tg.NewDeleteMessage(message.Chat.ID, message.MessageID),
				tg.BanChatSenderChatConfig{
					ChatID:       message.Chat.ID,
					SenderChatID: message.SenderChat.ID,
				},
			)
			if err != nil {
				newError("failed to ban channel messages from superchat").Base(err).WriteToLog()
			}

			return
		}
	}
	{
		// member check
		member, err := bot.GetChatMember(tg.GetChatMemberConfig{
			ChatConfigWithUser: tg.ChatConfigWithUser{
				ChatID: message.Chat.ID,
				UserID: message.From.ID,
			},
		})
		if err != nil {
			newError("failed to get chat member").Base(err).WriteToLog()
			return
		}
		if member.Status == "left" {
			send := tg.NewMessage(message.Chat.ID, "Join channel's chat group before leaving shitpost!")
			send.ReplyToMessageID = message.MessageID

			notice, err := bot.Send(send)
			if err != nil {
				newError("failed to send reply to message").Base(err).WriteToLog()
				return
			}
			err = bot.MustRequests(
				tg.NewDeleteMessage(message.Chat.ID, message.MessageID),
				tg.RestrictChatMemberConfig{
					ChatMemberConfig: tg.ChatMemberConfig{
						ChatID: message.Chat.ID,
						UserID: message.From.ID,
					},
					UntilDate: time.Now().Add(time.Minute).Unix(),
				},
			)
			if err != nil {
				newError("failed to ban messages from non-member").Base(err).WriteToLog()
				return
			}
			time.AfterFunc(10*time.Second, func() {
				bot.Request(tg.NewDeleteMessage(message.Chat.ID, notice.MessageID))
			})
		}
	}
	{
		// service messages check

		if message.NewChatMembers != nil || len(message.NewChatTitle) > 0 || len(message.NewChatPhoto) > 0 || message.DeleteChatPhoto {

			err := bot.MustRequest(tg.NewDeleteMessage(message.Chat.ID, message.MessageID))
			if err != nil {
				newError("failed to delete service message").Base(err).WriteToLog()
			}
		}
	}
}
