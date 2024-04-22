package telebot

import (
	"fmt"
	ss "golang-bot/sugar"
	"strings"

	tele "gopkg.in/telebot.v3"
)

var media_chan = make(chan *tele.Message, 10)

func create_Frames_n_Nets(ctx tele.Context) bool {
	// отправка списка рамок
	err := ctx.Send(MSG_FRAME, Markups["OnFrames"])
	if err != nil {
		return false
	}
	// отправка списка сеток
	err = ctx.Send(MSG_NET, Markups["OnNets"])
	if err != nil {
		return false
	}

	return true
}

func process_Frames_n_Nets(ctx tele.Context) error {
	user := TUser{}.Get(ctx.Sender().ID)
	order := TOrder{}.FromUser(user)
	order.ParseOptions(ctx.Data())
	user.Status = EXP_SIZES

	return send_OrderInfo(ctx)
}

func process_RequestMedia(ctx tele.Context) error {
	user := TUser{}.Get(ctx.Sender().ID)
	if user == nil {
		ss.Log("ERROR", "processRequestMedia", "Не удалось найти пользователя ни в кэше, ни в БД")
		return ctx.Send(ERR_RESTART)
	}
	user.Status = EXP_MEDIA
	answer := MSG_MEDIA
	if user.Phone == "" {
		answer += ",\nсодержащему <b>номер Вашего телефона и Имя.</b>"
	}
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

func send_OrderInfo(ctx tele.Context) error {
	user := TUser{}.Get(ctx.Sender().ID)
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
	if user.MessageOrder.Text == answer {
		return nil
	}
	msg, err := ctx.Bot().Edit(user.MessageOrder, answer, Markups["OnOrder"])
	if err == nil {
		user.MessageOrder = msg
	}

	return err
}

func validateOrder(ctx tele.Context, user *TUser) error {
	if user == nil {
		ss.Log("WARN", "validateOrder",
			fmt.Sprintf("Пользователь не %d не найдет. Ничего не делаем.", ctx.Sender().ID))
		return nil
	}
	if user.Status != EXP_CONTACT {
		ss.Log("WARN", "validateOrder",
			fmt.Sprintf("Пользователь %d пытается создать заказ, но мы этого не ждём", user.IDTele))
		return nil
	}
	if user.Phone == "" {
		ss.Log("INFO", "validateOrder",
			fmt.Sprintf("У нас нет контактов пользователя %d. Запрашиваем", user.IDTele))
		return ctx.Send(MSG_VALPHONE)
	}
	if user.Order == nil {
		user.Order = TOrder{}.New()
	}
	user.Order.IsPickup = strings.HasSuffix(ctx.Data(), "1")
	user.Order.DateTime = ss.GetDateTime()
	user.Order.CustomerID = user.IDTele
	go user.Order.AddToDb(user.IDTele)

	user.Status = EXP_START
	answer := answer_WillCallYou(user)
	Admin_BroadcastOrder(ctx, *user, ADMIN_GROUP, ORDER_NEW)

	return ctx.Send(answer)
}

func answer_WillCallYou(user *TUser) string {
	return fmt.Sprintf("Благодарю, %s!\n%s по номеру %s", user.FirstName, MSG_WILLCALL, user.Phone)
}

func isPrivate(ctx tele.Context) bool {
	user_id, chat_id := ctx.Sender().ID, ctx.Chat().ID
	my_name := fmt.Sprintf("@%s", ctx.Bot().Me.Username)
	if strings.HasPrefix(ctx.Message().Text, my_name) {
		ss.Log("CHECK", "isPrivate",
			fmt.Sprintf("(User: %d Chat: %d) Это приватное сообщение для бота", user_id, chat_id))
		return true
	}
	if ctx.Sender().ID == ctx.Chat().ID {
		ss.Log("CHECK", "isPrivate",
			fmt.Sprintf("(User: %d Chat: %d) Это приватное сообщение", user_id, chat_id))
		return true
	}
	ss.Log("CHECK", "isPrivate",
		fmt.Sprintf("(User: %d Chat: %d) Это сообщение в групповом чате", user_id, chat_id))
	return false
}

func broadcastOrder(ctx tele.Context, user TUser, chat *tele.Chat, title string) {
	answer := fmt.Sprintf("%s\n%s\n%s : %s\n - заказ звонка\n", title, ss.GetDateTime(), user.Phone, user.FirstName)
	order_info := user.Order.Display(true)
	if order_info != "" {
		answer = fmt.Sprintf("%s\n%s\n%s : %s\n%s- %s\n",
			title,
			user.Order.DateTime,
			user.Phone,
			user.FirstName,
			order_info,
			ss.Iif(user.Order.IsPickup, "самовывоз", "заказ замера"),
		)
	}
	msg, err := ctx.Bot().Send(chat, answer)
	if err != nil {
		ss.Log("ERROR", "Admin_BroadcastOrder", err.Error())
		return
	}
	OrderMessage{OrderID: user.Order.ID, Message: msg}.Add(chat.ID)
}

func broadcastMedia(ctx tele.Context, user TUser, chat *tele.Chat, title string) {
	answer := fmt.Sprintf("%s\n%s\n%s : %s", title, ss.GetDateTime(), user.Phone, user.FirstName)
	media := <-media_chan
	album := tele.Album{}
	if photo := media.Photo; photo != nil {
		photo.Caption = answer
		album = append(album, photo)
	}
	if video := media.Video; video != nil {
		video.Caption = answer
		album = append(album, video)
	}
	ctx.Bot().SendAlbum(chat, album, tele.ModeHTML)
	// _ = msg //OrderMessage{OrderID: user.Order.ID, Message: msg}.Add(chat.ID)
}
