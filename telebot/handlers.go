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
		"OnStart":  CreateButtonRows(bot, handleButton, append(ARR_START, "ADMIN"), "option"),
		"OnAdmin":  CreateButtonRows(bot, handleButton, ARR_ADMIN, "admin"),
		"OnFrames": CreateOptionCols(bot, handleButton, ARR_FRAMES, "frame"),
		"OnNets":   CreateOptionCols(bot, handleButton, ARR_NETS, "net"),
		"OnOrder":  CreateOptionCols(bot, handleButton, ARR_ORDER, "order"),
	}
	ss.Log("INFO", "PrepareMarkups", "Подготовка набора кнопок")
}

func Handle(ctx tele.Context, mode string) error {
	msg := ctx.Update().Message
	rct := ctx.Update().MessageReaction
	var sender *tele.User
	if msg != nil {
		sender = msg.Sender
	}
	if rct != nil {
		sender = rct.User
	}
	if sender == nil {
		ss.Log("ERROR", "Handle", "Не удалось получить данные пользователя")
		return nil
	}
	ss.Log("INFO", "Handle", fmt.Sprintf("Подключение пользователя %d", sender.ID))
	// идентификация юзера
	user := TUser{}.Get(sender.ID, sender.Username, sender.FirstName, sender.LastName)
	if !user.IsBanned {
		switch mode {
		case ON_START:
			handleStart(ctx, user)
		case ON_MESSAGE:
			handleMessage(ctx, user)
		case ON_REACTION:
			handleReaction(ctx)
		}
	}
	return nil
}

func handleStart(ctx tele.Context, user *TUser) {
	if user.IDTele != ctx.Chat().ID {
		ss.Log("WARN", "handleStart",
			fmt.Sprintf("Попытка стартовать из групового чата. Отказ! Пользователь %d", user.IDTele))
		ctx.Send(MSG_NOSTART)
		return
	}
	// формирование ответа на /start
	answer := fmt.Sprintf("👋  %s, %s!\n%s", ss.GenGreeting(), user.FirstName, MSG_START)
	msg, err := ctx.Bot().Send(ctx.Sender(), answer, Markups["OnStart"])
	if err != nil {
		ss.Log("ERROR", "handleStart",
			fmt.Sprintf("%d :: %v", user.IDTele, err.Error()))
		return
	}
	user.MessageLast = msg
	// сброс заказа и ссылки на сообщение с ценой и последнее сообщение
	user.Order = nil
	user.MessageOrder = nil
	user.Status = EXP_OPTION
	// обновляем время визита юзера
	user.UpdateVisit()
	// обновляем в кэшэ пользователей
	user.AddToCache()
}

func handleButton(ctx tele.Context) error {
	user := TUser{}.Get(ctx.Sender().ID)
	ss.Log(
		"INFO",
		"handleButton",
		fmt.Sprintf("Выбор пользователя %s:: %v", user.FirstName, ctx.Data()))

	switch data := ctx.Data(); {
	case data == BTN_ADMIN:
		Admin_ShowOptions(ctx)
	case strings.HasPrefix(data, "admin"):
		Admin_GetData(data, ctx)

	case data == BTN_CALCULATOR:
		if user.Status != EXP_OPTION {
			// ЗАЩИТА от спамминга
			// Если не ожидаем от пользователя этих действий - игнорим
			ss.Log("WARN", "handleButton", fmt.Sprintf("Неожидаемый выбор калькулятора от пользователя %d", user.IDTele))
			return nil
		}
		if create_Frames_n_Nets(ctx) {
			_ = ctx.Bot().Delete(user.MessageLast)
			user.MessageLast = nil
			user.Status = EXP_SIZES
		}
	// case data == BTN_SEND_MEDIA:
	// 	// TODO
	// 	return process_RequestMedia(ctx)
	case data == BTN_REQUEST_CALL:
		if user.Status != EXP_OPTION {
			// ЗАЩИТА от спамминга
			// Если не ожидаем от пользователя этих действий - игнорим
			ss.Log("WARN", "handleButton", fmt.Sprintf("Неожидаемый запрос обратного звонка от пользователя %d", user.IDTele))
			return nil
		}
		user.Status = EXP_CONTACT
		validateOrder(ctx, user)
	case strings.HasPrefix(data, "frame") || strings.HasPrefix(data, "net"):
		if user.Status != EXP_SIZES {
			// ЗАЩИТА от спамминга
			// Если не ожидаем от пользователя этих действий - игнорим
			ss.Log("WARN", "handleButton", fmt.Sprintf("Неожидаемый выбор сетки от пользователя %d", user.IDTele))
			return nil
		}
		process_Frames_n_Nets(ctx)
	case strings.HasPrefix(data, "order"):
		if user.Status != EXP_SIZES {
			// ЗАЩИТА от спамминга
			// Если не ожидаем от пользователя этих действий - игнорим
			ss.Log("WARN", "handleButton", fmt.Sprintf("Неожидаемый оформление заказа от пользователя %d", user.IDTele))
			return nil
		}
		user.Status = EXP_CONTACT
		validateOrder(ctx, user)
	}

	return nil
}

func handleMessage(ctx tele.Context, user *TUser) {
	msg := ctx.Message().Text
	if len(msg) > 64 {
		// ЗАЩИТА
		// Ограничене длины сообщения 64 символами
		msg = msg[:63]
	}
	ss.Log(
		"INFO",
		"handleMessage",
		fmt.Sprintf("Сообщение пользователя %s:: %s = %s", user.UserName, msg, user.Status))

	switch user.Status {
	case EXP_CONTACT:
		ss.Log("INFO", "handleMessage", "Обработка контакта")
		answer := process_Contact(user, msg)
		if answer != "" {
			if answer != MSG_ERRPHONE {
				validateOrder(ctx, user)
				return
			}
			ctx.Send(answer)
		}
	case EXP_SIZES:
		ss.Log("INFO", "handleMessage", "Обработка размеров")
		order := TOrder{}.FromUser(user)
		if order.ParseSizes(msg) {
			_ = ctx.Delete()
			send_OrderInfo(ctx)
		}
	}
	ss.Log("WARN", "handleMessage",
		fmt.Sprintf("Не обрабатываемое сообщение от пользователя %d:: %s", ctx.Sender().ID, ctx.Message().Text),
	)
	ctx.Delete()
}
func handleReaction(ctx tele.Context) {
	reaction := ctx.Update().MessageReaction
	ss.Log("INFO", "handleReaction", fmt.Sprintf("Эмодзи от пользователя %d", reaction.User.ID))
	user := TUser{}.Get(reaction.User.ID)
	if user != nil && user.IsAdmin {
		go Admin_ReactToOrder(ctx, *reaction)
	}
}

func HandleMedia(ctx tele.Context) error {
	user := TUser{}.Get(ctx.Sender().ID)
	if user == nil {
		return nil
	}
	if user.Status != EXP_MEDIA {
		// ЗАЩИТА от спамминга
		// Если не ожидаем от пользователя 'этих действий - игнорим
		return nil
	}
	message := ctx.Message()
	if message.Video == nil && message.Photo == nil {
		return nil
	}
	media_chan <- ctx.Message()
	contact := process_Contact(user, ctx.Message().Text)
	if contact == "" && user.Phone == "" {
		user.Status = EXP_CONTACT
		return ctx.Send("Укажите, пожалуйста, <b>номер Вашего телефона и Имя</b>")
	}
	broadcastMedia(ctx, *user, ADMIN_GROUP, ORDER_MEDIA)

	return ctx.Send(answer_WillCallYou(user))
}
