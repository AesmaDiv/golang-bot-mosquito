package main

import (
	ss "golang-bot/sugar"
	tb "golang-bot/telebot"
	"os"
	"time"

	env "github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
)

func main() {
	// пробую подключиться к БД
	if !tb.ConnectToDb() {
		return
	}
	// отключиться при завершении функции
	defer tb.DisconnectFromDb()

	ss.Log("INFO", "TELEBOT", "Загрузка переменных окружения..")
	err := env.Load()
	ss.CheckError(err)

	ss.Log("INFO", "TELEBOT", "Создание нового бота..")
	token := os.Getenv("TOKEN")
	pref := tele.Settings{
		Token: token,
		Poller: &tele.LongPoller{
			Timeout:      1 * time.Second,
			LastUpdateID: -2,
			Limit:        100,
			AllowedUpdates: []string{
				"message",
				"callback_query",
				"message_reaction",
				"message_reaction_count",
			},
		},
		ParseMode: tele.ModeHTML,
	}
	bot, err := tele.NewBot(pref)
	ss.CheckError(err)

	bot.Handle("/start", func(ctx tele.Context) error { return tb.Handle(ctx, tb.ON_START) })
	bot.Handle(tele.OnText, func(ctx tele.Context) error { return tb.Handle(ctx, tb.ON_MESSAGE) })
	bot.Handle(tele.OnReaction, func(ctx tele.Context) error { return tb.Handle(ctx, tb.ON_REACTION) })
	// bot.Handle(tele.OnMedia, tb.HandleMedia)

	tb.PrepareMarkups(bot)
	ss.Log("INFO", "TELEBOT", "Запуск бота..")
	bot.Start()

	ss.Log("INFO", "TELEBOT", "Бот завершил работу.")
}
