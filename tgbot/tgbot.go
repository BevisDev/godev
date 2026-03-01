package tgbot

import (
	"context"
	"sync"
	"time"

	"github.com/BevisDev/godev/utils/console"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgBot struct {
	*options
	bot           *tgbotapi.BotAPI
	logger        *console.Logger
	mu            sync.RWMutex
	sessionExpiry map[int64]time.Time
}

func New(token string, opts ...Option) (*TgBot, error) {
	opt := withDefaults()
	for _, o := range opts {
		o(opt)
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	b := &TgBot{
		options:       opt,
		bot:           bot,
		logger:        console.New("tgbot"),
		sessionExpiry: make(map[int64]time.Time),
	}

	b.logger.Info("started successfully")
	return b, nil
}

func (t *TgBot) BotAPI() *tgbotapi.BotAPI {
	return t.bot
}

func (t *TgBot) Send(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := t.bot.Send(msg); err != nil {
		return err
	}
	return nil
}

func (t *TgBot) StartLongPolling(
	ctx context.Context,
	handler func(tgbotapi.Update),
) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := t.bot.GetUpdatesChan(u)
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			handler(update)
		}
	}
}

func (t *TgBot) NewSession(chatID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sessionExpiry[chatID] = time.Now().Add(t.sessionDuration)
}

func (t *TgBot) IsSessionActive(chatID int64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	exp, ok := t.sessionExpiry[chatID]
	if !ok {
		return false
	}

	if time.Now().After(exp) {
		delete(t.sessionExpiry, chatID)
		return false
	}
	return true
}
