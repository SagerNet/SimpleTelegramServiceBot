package main

import tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func (bot *Service) MustRequest(request tg.Chattable) error {
	response, err := bot.Request(request)
	if err != nil {
		return err
	}
	if !response.Ok {
		return newError("error ", response.ErrorCode, ": ", response.Description)
	}
	return nil
}

func (bot *Service) MustRequests(requests ...tg.Chattable) error {
	for _, request := range requests {
		if err := bot.MustRequest(request); err != nil {
			return err
		}
	}
	return nil
}
