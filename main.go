package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "c", "config.json", "path to config file")
}

type Config struct {
	BotToken string `json:"botToken"`
}

func main() {
	flag.Parse()
	configBytes, err := os.ReadFile(configPath)
	common.Must(err)
	config := &Config{}
	common.Must(json.Unmarshal(configBytes, config))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b, err := bot.New(config.BotToken, bot.WithDefaultHandler(handleUpdate))
	if err != nil {
		panic(err)
	}

	b.Start(ctx)
}

func handleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil {
		onNewMessage(ctx, b, update.Message)
	}
}

func onNewMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	if message.Chat.Type != models.ChatTypeSupergroup {
		return
	}
	{
		// channel message check
		if message.SenderChat != nil {
			chat, err := b.GetChat(ctx, &bot.GetChatParams{
				ChatID: message.Chat.ID,
			})
			if err != nil {
				log.Error(E.Cause(err, "failed to get chat"))
				return
			}

			if message.SenderChat.ID == message.Chat.ID || message.SenderChat.ID == chat.LinkedChatID {
				return
			}

			_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
				ChatID:    message.Chat.ID,
				MessageID: message.ID,
			})
			if err != nil {
				log.Error(E.Cause(err, "failed to delete channel messages from superchat"))
			}

			_, err = b.BanChatSenderChat(ctx, &bot.BanChatSenderChatParams{
				ChatID:       message.Chat.ID,
				SenderChatID: int(message.SenderChat.ID),
			})
			if err != nil {
				log.Error(E.Cause(err, "failed to ban channel messages from superchat"))
				return
			}

			return
		}
	}
	{
		// member check
		member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
			ChatID: message.Chat.ID,
			UserID: message.From.ID,
		})
		if err != nil {
			log.Error(E.Cause(err, "failed to get chat member"))
			return
		}
		if member.Type == models.ChatMemberTypeLeft {
			sentMessage, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ReplyParameters: &models.ReplyParameters{
					ChatID:    message.Chat.ID,
					MessageID: message.ID,
				},
			})
			if err != nil {
				log.Error(E.Cause(err, "failed to send reply to message"))
				return
			}
			_, err = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
				ChatID:    message.Chat.ID,
				MessageID: message.ID,
			})
			if err != nil {
				log.Error(E.Cause(err, "failed to delete non-member message"))
				return
			}
			_, err = b.RestrictChatMember(ctx, &bot.RestrictChatMemberParams{
				ChatID:    message.Chat.ID,
				UserID:    message.From.ID,
				UntilDate: int(time.Now().Add(time.Minute).Unix()),
			})
			if err != nil {
				log.Error(E.Cause(err, "failed to restrict chat member"))
				return
			}
			time.AfterFunc(10*time.Second, func() {
				_, _ = b.DeleteMessage(ctx, &bot.DeleteMessageParams{
					ChatID:    message.Chat.ID,
					MessageID: sentMessage.ID,
				})
			})
		}
	}
	{
		// service messages check

		if message.NewChatMembers != nil ||
			len(message.NewChatTitle) > 0 ||
			len(message.NewChatPhoto) > 0 ||
			message.DeleteChatPhoto ||
			message.PinnedMessage.Message != nil ||
			message.BoostAdded != nil {
			_, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{
				ChatID:    message.Chat.ID,
				MessageID: message.ID,
			})
			if err != nil {
				log.Error(E.Cause(err, "failed to delete service message"))
			}
		}
	}
}
