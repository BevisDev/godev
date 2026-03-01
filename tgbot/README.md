# tgbot

Thin wrapper around [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) with session tracking and configurable options. Uses the project's `console` logger and supports functional options for construction.

## Installation

Use as a module within this repo:

```go
import "github.com/BevisDev/godev/tgbot"
```

## Quick start

```go
bot, err := tgbot.New(os.Getenv("TELEGRAM_BOT_TOKEN"))
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()
bot.StartLongPolling(ctx, func(update tgbotapi.Update) {
    if update.Message != nil {
        bot.NewSession(update.Message.Chat.ID)
        _ = bot.Send(update.Message.Chat.ID, "Hello!")
    }
})
```

## Options

- **WithSessionDuration(d time.Duration)** — How long a chat session stays active. Default: 1 hour. Pass `d > 0` to override.

Example:

```go
bot, err := tgbot.New(token, tgbot.WithSessionDuration(30*time.Minute))
```

## API overview

| Symbol            | Description |
|-------------------|-------------|
| `New(token, opts...)` | Create a bot with optional config. |
| `TgBot.BotAPI()`  | Underlying `*tgbotapi.BotAPI` for advanced use. |
| `TgBot.Send(chatID, text)` | Send a plain text message to a chat. |
| `TgBot.Reply(chatID, replyToMessageID, text)` | Send a message as a reply to another message. |
| `TgBot.SendMarkdown(chatID, text)` | Send a message with Markdown parse mode. |
| `TgBot.SendHTML(chatID, text)` | Send a message with HTML parse mode. |
| `TgBot.EditMessageText(chatID, messageID, text)` | Edit the text of an existing message. |
| `TgBot.AnswerCallback(callbackQueryID, text)` | Answer a callback query (e.g. from inline keyboard); optional alert text. |
| `tgbot.ChatID(update)` | Get chat ID from an update (message or callback); returns `(int64, bool)`. |
| `tgbot.MessageID(update)` | Get message ID from an update for Reply/EditMessageText; returns `(int, bool)`. |
| `TgBot.StartLongPolling(ctx, handler)` | Long-poll updates until `ctx` is done; calls handler per update. |
| `TgBot.NewSession(chatID)` | Mark chat as having an active session; resets expiry. |
| `TgBot.IsSessionActive(chatID)` | Report whether the chat has a non-expired session. |

## Session behaviour

Sessions are per-chat and time-based. Call `NewSession(chatID)` when a user starts or continues an interaction; use `IsSessionActive(chatID)` to decide if you should resume or start fresh. Expired entries are removed when checked.
