package telebot

import (
	"fmt"
	ss "golang-bot/sugar"
	"strings"

	tele "gopkg.in/telebot.v3"
)

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

func process_RequestCall(ctx tele.Context) error {
	answer, user := "", TUser{}.Get(helper, ctx.Sender().ID)
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
