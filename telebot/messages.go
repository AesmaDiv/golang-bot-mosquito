package telebot

import (
	"fmt"
	ss "golang-bot/sugar"

	tele "gopkg.in/telebot.v3"
)

type OrderMessage struct {
	OrderID int
	Message *tele.Message
}

var refs_to_msgs = make(map[int64][]OrderMessage)

func (om OrderMessage) Add(id_chat int64) {
	if om.OrderID == 0 || om.Message == nil {
		return
	}
	refs_to_msgs[id_chat] = append(refs_to_msgs[id_chat], om)
	ss.Log(
		"SUCCESS",
		"Admin_RegisterOrderMessage",
		fmt.Sprintf("Заказ %d успешно привязан к сообщению %d", om.OrderID, om.Message.ID),
	)
}

func Message_GetOrder(id_msg int, id_chat int64) int {
	if items, found := refs_to_msgs[id_chat]; found {
		for _, item := range items {
			if item.Message.ID == id_msg {
				return item.OrderID
			}
		}
	}

	return 0
}
func Message_Delete(bot *tele.Bot, id_msg int, id_chat int64) {
	if items, found := refs_to_msgs[id_chat]; found {
		for i, item := range items {
			if item.Message.ID == id_msg {
				refs_to_msgs[id_chat] = append(items[:i], items[i+1:]...)
				bot.Delete(item.Message)
				break
			}
		}
	}
}

func (om OrderMessage) Clean(id_chat int64) {
	delete(refs_to_msgs, id_chat)
}
