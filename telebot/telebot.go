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
		Host:   "localhost",
		Port:   5432,
		User:   "postgres",
		Pswd:   "123",
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
	ss.Log("INFO", "handleStart", fmt.Sprintf("Подключение пользователя %v", *user))
	// формирование ответа на /start
	// приветствие
	answer := fmt.Sprintf("👋  %s, %s!\n%s", ss.GenGreeting(), user.FirstName, MSG_START)
	// начальные кнопки
	// buttons := ON_START[:]
	// // если пользователь админ - добавить админку
	// // if user.IsAdmin {
	// buttons = append(buttons, "АДМИНКА")
	// // }
	// markup := CreateButtonRows(ctx.Bot(), handleButton, buttons, "option")

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
		return send_RequestContact(ctx)

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
				validateOrder(ctx)
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

func create_Frames_n_Nets(ctx tele.Context) error {
	// отправка списка рамок
	err := ctx.Send(MSG_FRAME, Markups["OnFrames"])
	if err != nil {
		return err
	}
	// отправка списка сеток
	return ctx.Send(MSG_NET, Markups["OnNets"])
}

func process_Frames_n_Nets(ctx tele.Context) error {
	user := TUser{}.Get(helper, ctx.Sender().ID)
	order := TOrder{}.FromUser(user)
	order.ParseOptions(ctx.Data())
	user.Status = EXP_SIZES

	return send_OrderInfo(ctx)
}

func send_RequestContact(ctx tele.Context) error {
	user := TUser{}.Get(helper, ctx.Sender().ID)
	answer := MSG_ASKPHONE
	if user.Phone != "" {
		answer = fmt.Sprintf("Благодарю за обращение, %s!\n%s", user.FirstName, MSG_WILLCALL)
	}
	user.Status = EXP_CONTACT

	return ctx.Send(answer)
}

func send_OrderInfo(ctx tele.Context) error {
	user := TUser{}.Get(helper, ctx.Sender().ID)
	order := TOrder{}.FromUser(user)
	answer := order.Display(false)
	if answer == "" {
		return nil
	}
	if user.MessageOrder == nil {
		msg, err := ctx.Bot().Send(ctx.Recipient(), answer, Markups["OnOrder"])
		user.MessageOrder = msg
		return err
	}
	msg, err := ctx.Bot().Edit(user.MessageOrder, answer, Markups["OnOrder"])
	if err == nil {
		user.MessageOrder = msg
	}

	return err
}

func send_RequestMedia(ctx tele.Context) error {
	return ctx.Send(MSG_MEDIA)
}

func validateOrder(ctx tele.Context) error {
	user := TUser{}.Get(helper, ctx.Sender().ID)
	if user.Phone == "" {
		user.Status = EXP_CONTACT
		return ctx.Send(MSG_ASKPHONE)
	}
	user.Order.IsPickup = strings.HasSuffix(ctx.Data(), "1")
	user.Order.DateTime = ss.GetDateTime()
	user.Order.CustomerID = user.TeleID
	go func() {
		user.Order.AddToDb(user.TeleID)
		Admin_BroadcastOrder(ctx, *user, true, ORDER_NEW)
	}()
	answer := answer_WillCallYou(user)

	return ctx.Send(answer)
}

func process_Contact(user *TUser, msg string) string {
	user.ParseContact(msg)
	// если телефон мы не получили
	if user.Phone == "" {
		return MSG_ERRPHONE
	}
	user.DBUpdate_Contact(helper)
	// формируем ответ для пользователя
	answer := answer_WillCallYou(user)
	return answer
}

func answer_WillCallYou(user *TUser) string {
	return fmt.Sprintf("Благодарю, %s!\n%s по номеру %s", user.FirstName, MSG_WILLCALL, user.Phone)
}

func isPrivate(ctx tele.Context) bool {
	ss.Log("CHECK", "USER", ss.ToString(ctx.Sender().ID))
	ss.Log("CHECK", "GROUP", ss.ToString(ctx.Chat().ID))
	my_name := fmt.Sprintf("@%s", ctx.Bot().Me.Username)
	if strings.HasPrefix(ctx.Message().Text, my_name) {
		return true
	}
	if ctx.Sender().ID == ctx.Chat().ID {
		return true
	}
	return false
}
