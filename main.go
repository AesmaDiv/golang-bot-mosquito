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
			Timeout: 1 * time.Second,
			AllowedUpdates: []string{
				"message",
				"edited_message",
				"channel_post",
				"edited_channel_post",
				"inline_query",
				"chosen_inline_result",
				"callback_query",
				"shipping_query",
				"pre_checkout_query",
				"poll",
				"poll_answer",
				"message_reaction",
				"message_reaction_count",
			},
		},
		ParseMode: tele.ModeHTML,
	}
	bot, err := tele.NewBot(pref)
	ss.CheckError(err)

	bot.Handle("/start", tb.HandleStart)
	bot.Handle(tele.OnText, tb.HandleMessage)
	bot.Handle(tele.OnReaction, tb.HandleReaction)
	bot.Handle(tele.OnMedia, tb.HandleMedia)

	tb.PrepareMarkups(bot)
	// bot.Handle(tele.OnAddedToGroup, tb.HandleGroup)
	ss.Log("INFO", "TELEBOT", "Запуск бота..")
	bot.Start()

	ss.Log("INFO", "TELEBOT", "Бот завершил работу.")
}
