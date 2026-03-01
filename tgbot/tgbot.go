// Package tgbot provides a thin wrapper around telegram-bot-api with session tracking and configurable options.
package tgbot

import (
	"context"
	"sync"
	"time"

	"github.com/BevisDev/godev/utils/console"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TgBot wraps a Telegram Bot API client with session expiry and logging.
type TgBot struct {
	*options
	bot           *tgbotapi.BotAPI
	logger        *console.Logger
	mu            sync.RWMutex
	sessionExpiry map[int64]time.Time
}

// New creates a TgBot with the given token and optional configuration.
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

// BotAPI returns the underlying telegram-bot-api client for advanced use.
func (t *TgBot) BotAPI() *tgbotapi.BotAPI {
	return t.bot
}

// Send sends a text message to the given chat.
func (t *TgBot) Send(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := t.bot.Send(msg); err != nil {
		return err
	}
	return nil
}

// Reply sends a text message as a reply to the given message in the same chat.
func (t *TgBot) Reply(chatID int64, replyToMessageID int, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = replyToMessageID
	if _, err := t.bot.Send(msg); err != nil {
		return err
	}
	return nil
}

// SendMarkdown sends a message with Markdown parse mode (e.g. *bold*, _italic_, [text](url)).
func (t *TgBot) SendMarkdown(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if _, err := t.bot.Send(msg); err != nil {
		return err
	}
	return nil
}

// SendHTML sends a message with HTML parse mode (e.g. <b>bold</b>, <a href="">link</a>).
func (t *TgBot) SendHTML(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	if _, err := t.bot.Send(msg); err != nil {
		return err
	}
	return nil
}

// EditMessageText edits the text of an existing message in the given chat.
func (t *TgBot) EditMessageText(chatID int64, messageID int, text string) error {
	cfg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	_, err := t.bot.Send(cfg)
	return err
}

// AnswerCallback answers a callback query (e.g. from inline keyboard). Optional text shows an alert to the user.
func (t *TgBot) AnswerCallback(callbackQueryID, text string) error {
	cfg := tgbotapi.NewCallback(callbackQueryID, text)
	_, err := t.bot.Request(cfg)
	return err
}

// ChatID returns the chat ID from an update (message or callback query). The second return is false if neither is present (e.g. inline callback with no message).
func ChatID(u tgbotapi.Update) (int64, bool) {
	if u.Message != nil {
		return u.Message.Chat.ID, true
	}
	if u.CallbackQuery != nil && u.CallbackQuery.Message != nil {
		return u.CallbackQuery.Message.Chat.ID, true
	}
	return 0, false
}

// MessageID returns the message ID from an update, for use with Reply or EditMessageText. The second return is false if no message is present.
func MessageID(u tgbotapi.Update) (int, bool) {
	if u.Message != nil {
		return u.Message.MessageID, true
	}
	if u.CallbackQuery != nil && u.CallbackQuery.Message != nil {
		return u.CallbackQuery.Message.MessageID, true
	}
	return 0, false
}

// StartLongPolling runs the update loop until ctx is done, calling handler for each update.
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

// NewSession marks a chat as having an active session and resets its expiry.
func (t *TgBot) NewSession(chatID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.sessionExpiry[chatID] = time.Now().Add(t.sessionDuration)
}

// IsSessionActive reports whether the chat has a non-expired session.
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
