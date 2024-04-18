package telebot

import (
	"fmt"
	ss "golang-bot/sugar"
	"strings"

	tele "gopkg.in/telebot.v3"
)

var media_chan = make(chan *tele.Message, 10)

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

func process_RequestCall(ctx tele.Context) error {
	answer, user := "", TUser{}.Get(ctx.Sender().ID)
	if user.Phone == "" {
		answer = MSG_ASKPHONE
		user.Status = EXP_CONTACT
	} else {
		answer = answer_WillCallYou(user)
		user.Status = ""
	}

	return ctx.Send(answer)
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
	msg, err := ctx.Bot().Edit(user.MessageOrder, answer, Markups["OnOrder"])
	if err == nil {
		user.MessageOrder = msg
	}

	return err
}

func validateOrder(ctx tele.Context) error {
	user := TUser{}.Get(ctx.Sender().ID)
	if user.Phone == "" {
		user.Status = EXP_CONTACT
		return ctx.Send(MSG_VALPHONE)
	}
	if user.Order != nil {
		user.Order.IsPickup = strings.HasSuffix(ctx.Data(), "1")
		user.Order.DateTime = ss.GetDateTime()
		user.Order.CustomerID = user.TeleID
		go func() {
			user.Order.AddToDb(user.TeleID)
			Admin_BroadcastOrder(ctx, *user, ADMIN_GROUP, ORDER_NEW)
		}()
	}
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

func broadcastMessage(ctx tele.Context, user TUser, chat *tele.Chat, title string) {
	user_info := parseUserInfo(map[string]any{
		"datetime": user.Order.DateTime,
		"phone":    user.Phone,
		"fname":    user.FirstName,
	})
	order_info := user.Order.Display(true)
	answer := fmt.Sprintf("%s\n%s%s- %s\n",
		title,
		user_info,
		order_info,
		ss.Iif(user.Order.IsPickup, "самовывоз", "заказ замера"),
	)
	msg, err := ctx.Bot().Send(chat, answer)
	if err != nil {
		ss.Log("ERROR", "Admin_BroadcastOrder", err.Error())
		return
	}
	OrderMessage{OrderID: user.Order.ID, Message: msg}.Add(chat.ID)
}

func broadcastMedia(ctx tele.Context, user TUser, chat *tele.Chat, title string) {
	user_info := parseUserInfo(map[string]any{
		"datetime": ss.GetDateTime(),
		"phone":    user.Phone,
		"fname":    user.FirstName,
	})
	answer := fmt.Sprintf("%s\n%s\n", title, user_info)
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
