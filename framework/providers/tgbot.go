package providers

import (
	"context"

	"github.com/BevisDev/godev/tgbot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgBotProvider struct {
	cfg     *tgbot.Config
	opts    []tgbot.Option
	handler func(tgbotapi.Update)
	bot     *tgbot.TgBot
}

func NewTgBotProvider(cfg *tgbot.Config, handler func(tgbotapi.Update), opts ...tgbot.Option) *TgBotProvider {
	return &TgBotProvider{
		cfg:     cfg,
		opts:    opts,
		handler: handler,
	}
}

func (p *TgBotProvider) Init(ctx context.Context) error {
	_ = ctx
	bot, err := tgbot.New(p.cfg, p.opts...)
	if err != nil {
		return err
	}
	p.bot = bot
	return nil
}

func (p *TgBotProvider) Start(ctx context.Context) error {
	if p.bot != nil && p.handler != nil {
		go p.bot.StartLongPolling(ctx, p.handler)
	}
	return nil
}

func (p *TgBotProvider) Stop(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *TgBotProvider) Bot() *tgbot.TgBot {
	return p.bot
}
