package telebot

import (
	"fmt"
	db "golang-bot/database"
	ss "golang-bot/sugar"
	"strings"

	tele "gopkg.in/telebot.v3"
)

var helper = &db.HelperPostgres{}

func ConnectToDb() bool {
	ss.Log("INFO", "POSTGRES", "Подключение к БД..")
	result := helper.Connect(db.ConnectionParams{
		Host:   "185.198.152.151",
		Port:   5432,
		User:   "postgres",
		Pswd:   "aesma123div",
		Dbname: "telegram",
	})
	ss.Log(
		ss.Iif(result, "SUCCESS", "ERROR"),
		"ConnectToDB",
		ss.Iif(result, "БД подключена успешно..", "Не удалось подключиться к БД"),
	)

	return result
}

func DisconnectFromDb() {
	helper.Disconnect()
}

func PrepareMarkups(bot *tele.Bot) {
	Markups = map[string]*tele.ReplyMarkup{
		"OnStart":  CreateButtonRows(bot, handleButton, append(ON_START, "ADMIN"), "option"),
		"OnAdmin":  CreateButtonRows(bot, handleButton, ON_ADMIN, "admin"),
		"OnFrames": CreateOptionCols(bot, handleButton, FRAMES, "frame"),
		"OnNets":   CreateOptionCols(bot, handleButton, NETS, "net"),
		"OnOrder":  CreateOptionCols(bot, handleButton, ON_ORDER, "order"),
	}
	ss.Log("INFO", "PrepareMarkups", "Подготовка набора кнопок")
}

func HandleStart(ctx tele.Context) error {
	// идентификация юзера
	sender := ctx.Sender()
	if sender.ID != ctx.Chat().ID {
		return ctx.Send(MSG_NOSTART)
	}
	user := TUser{}.Get(helper, sender.ID)
	if user == nil {
		// если не нашли - добавляем
		user = TUser{}.New(sender.ID, ctx.Chat().ID, sender.Username, sender.FirstName)
		user.AddToDb(helper)
	}
	// сброс заказа и ссылки на сообщение с ценой и последнее сообщение
	user.Order = nil
	user.MessageOrder = nil
	user.MessageLast = nil
	// обновляем время визита юзера
	user.DBUpdate_Visit(helper)
	// обновляем в кэшэ пользователей
	user.AddToCache()
	ss.Log("INFO", "handleStart", fmt.Sprintf("Подключение пользователя %d:: %s", user.TeleID, user.FirstName))
	// формирование ответа на /start
	// приветствие
	answer := fmt.Sprintf("👋  %s, %s!\n%s", ss.GenGreeting(), user.FirstName, MSG_START)

	return ctx.Send(answer, Markups["OnStart"])
}

func handleButton(ctx tele.Context) error {
	ss.Log(
		"INFO",
		"handleButton",
		fmt.Sprintf("Выбор пользователя %s:: %v", ctx.Sender().Username, ctx.Data()))

	switch data := ctx.Data(); {
	case data == BTN_ADMIN:
		return ctx.Send(MSG_ADMIN, Markups["OnAdmin"])
	case strings.HasPrefix(data, "admin"):
		Admin_GetData(helper, data, ctx)

	case data == BTN_SHOW_OPTIONS:
		return create_Frames_n_Nets(ctx)
	case data == BTN_SEND_MEDIA:
		// TODO
		return send_RequestMedia(ctx)
	case data == BTN_REQUEST_CALL:
		return process_RequestCall(ctx)

	case strings.HasPrefix(data, "frame") || strings.HasPrefix(data, "net"):
		return process_Frames_n_Nets(ctx)
	case strings.HasPrefix(data, "order"):
		return validateOrder(ctx)
	}

	return nil
}

func HandleMessage(ctx tele.Context) error {
	ss.Log(
		"INFO",
		"handleMessage",
		fmt.Sprintf("Сообщение пользователя %s:: %s\n", ctx.Sender().Username, ctx.Message().Text))

	if !isPrivate(ctx) {
		return handleMessage_Group(ctx)
	}
	user := TUser{}.Get(helper, ctx.Sender().ID)
	if user.IsAdmin {
		return handleMessage_Admins(ctx, user)
	}
	return handleMessage_Users(ctx, user)
}

func HandleReaction(ctx tele.Context) error {
	reaction := ctx.Update().MessageReaction
	go Admin_ReactToOrder(ctx, *reaction)
	return nil
}

func HandleGroup(ctx tele.Context) error {
	// ADMIN_GROUP = ctx.Chat()
	return nil
}

func handleMessage_Group(ctx tele.Context) error {
	msg := strings.ToLower(ctx.Message().Text)
	if strings.Contains(msg, "заказ") {
		Admin_GetOrders(ctx, true, true)
	}
	return nil
}

func handleMessage_Users(ctx tele.Context, user *TUser) error {
	switch user.Status {
	case EXP_CONTACT:
		ss.Log("INFO", "handleMessage", "Обработка контакта")
		answer := process_Contact(user, ctx.Message().Text)
		if answer != "" {
			if answer != MSG_ERRPHONE {
				return validateOrder(ctx)
				//Order_AddToDb(user.Order, user.TeleID)
				// AddOrder(user)
			}
			return ctx.Send(answer)
		}
	case EXP_SIZES:
		ss.Log("INFO", "handleMessage", "Обработка размеров")
		order := TOrder{}.FromUser(user)
		order.ParseSizes(ctx.Message().Text)
		_ = ctx.Delete()

		return send_OrderInfo(ctx)
	}

	ss.Log("ERROR", "handleMessage", "Некорректный запрос")
	return ctx.Send(ctx.Message().Text)
}

func handleMessage_Admins(ctx tele.Context, user *TUser) error {
	_ = user
	// ctx.Bot().Send(ADMIN_GROUP, fmt.Sprintf("User %s:: %s", user.UserName, ctx.Message().Text))
	return ctx.Send("Как прикажешь, Повелитель")
}
