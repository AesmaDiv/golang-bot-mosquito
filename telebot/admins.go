package telebot

import (
	"fmt"
	db "golang-bot/database"
	ss "golang-bot/sugar"

	tele "gopkg.in/telebot.v3"
)

var ADMIN_GROUP = &tele.Chat{
	ID:   -1002059521153,
	Type: "group",
}

func Admin_CreatePanel(ctx tele.Context) error {
	markup := CreateButtonRows(ctx.Bot(), handleButton, ON_ADMIN, "admin")
	answer := MSG_ADMIN

	return ctx.Send(answer, markup)
}

func Admin_GetData(helper db.Helper, data string, ctx tele.Context) error {
	//var answer string
	switch data {
	case ADMIN_CUSTOMERS:
		go Admin_GetUsers(ctx, false)
	case ADMIN_MY_ORDERS:
		go Admin_GetOrders(ctx, false, false)
		return nil
	case ADMIN_FREE_ORDERS:
		go Admin_GetOrders(ctx, true, false)
		return nil
		// answer = requestOrders(helper, "Where is_closed=false;")
	}

	return nil //ctx.Send(answer, tele.ModeHTML)
}

func Admin_GetUsers(ctx tele.Context, is_admins bool) {
	customers := helper.Select(
		"users",
		[]string{"uname", "fname", "lname", "phone", "visit"},
		map[string]any{"is_admin": is_admins},
	)
	answer := "Список пользователей пуст."
	if len(customers) > 0 {
		answer = ss.Iif(is_admins, "======АДМИНЫ======\n", "======КЛИЕНТЫ======\n")
		for _, customer := range customers {
			user := TUser{}.FromMap(customer)
			answer += user.Display()
		}
	}
	ctx.Send(answer, tele.ModeHTML)
}

func Admin_GetOrders(ctx tele.Context, free, to_group bool) {
	where := fmt.Sprintf("WHERE o.id_worker = %d", ctx.Sender().ID)
	title := ORDER_YOUR
	if free {
		where = "WHERE o.id_worker < 1 OR o.id_worker IS NULL"
		title = ORDER_FREE
	}
	items := TOrder{}.GetOrdersFull(where)
	for _, item := range items {
		user := parseUserWithOrder(item)
		Admin_BroadcastOrder(ctx, user, to_group, title)
	}
}

func Admin_BroadcastOrder(ctx tele.Context, user TUser, to_group bool, title string) {
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
	var chat *tele.Chat = ADMIN_GROUP
	if !to_group {
		chat = ctx.Chat()
	}
	msg, err := ctx.Bot().Send(chat, answer, tele.ModeHTML)
	if err != nil {
		ss.Log("ERROR", "Admin_BroadcastOrder", err.Error())
		return
	}
	OrderMessage{OrderID: user.Order.ID, Message: msg}.Add(chat.ID)
}

func Admin_RegisterOrderMessage(id_tele int64, id_order int, id_message int) {
	if id_message == 0 || id_order == 0 {
		return
	}
}

func Admin_ReactToOrder(ctx tele.Context, react tele.MessageReaction) {
	id_worker := react.User.ID
	id_order := Message_GetOrder(react.MessageID, react.Chat.ID)
	if id_order == 0 {
		ss.Log(
			"ERROR", "Admin_ReactToOrder",
			fmt.Sprintf("Не получилось найти заказ %d в сообщениях", id_order),
		)
		return
	}
	items := TOrder{}.GetOrdersFull(fmt.Sprintf("WHERE o.id=%d", id_order))
	if len(items) == 0 {
		ss.Log(
			"ERROR", "Admin_ReactToOrder",
			fmt.Sprintf("Не удалось найти заказ %d в БД", id_order),
		)
		return
	}
	user := parseUserWithOrder(items[0])
	switch worker := user.Order.WorkerID; {
	case worker < 1:
		_ = user.Order.UpdateWorker(id_order, id_worker)
		Message_Delete(ctx.Bot(), react.MessageID, react.Chat.ID)
		ss.Log("SUCCESS", "Admin_ReactToOrder", fmt.Sprintf("Заказ %d привязан исполнителю %d", id_order, id_worker))
	case worker == id_worker:
		_ = user.Order.UpdateWorker(id_order, 0)
		Message_Delete(ctx.Bot(), react.MessageID, react.Chat.ID)
		Admin_BroadcastOrder(ctx, user, true, ORDER_FREE)
		ss.Log("SUCCESS", "Admin_ReactToOrder", fmt.Sprintf("Заказ %d теперь свободен", id_order))
	}
}

func parseUserInfo(items map[string]any) string {
	return fmt.Sprintf("%s\n%s: <b>%s</b>\n",
		items["datetime"],
		items["phone"],
		items["fname"],
	)
}

func parseUserWithOrder(item map[string]any) TUser {
	return TUser{
		FirstName: ss.ToString(item["fname"]),
		Phone:     ss.ToString(item["phone"]),
		Order:     TOrder{}.FromMap(item),
	}
}
