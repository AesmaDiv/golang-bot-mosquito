package telebot

import (
	"fmt"
	ss "golang-bot/sugar"

	tele "gopkg.in/telebot.v3"
)

func CreateMarkup(bot *tele.Bot, f tele.HandlerFunc) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	l := ss.GetMaxLength(FRAMES, NETS)
	var rows []tele.Row
	for i := 0; i < l; i++ {
		var btn1, btn2 tele.Btn
		if btn1 = markup.Data(" ", fmt.Sprintf("dull_%d", i)); i < len(FRAMES) {
			name := fmt.Sprintf("frame_%d", i)
			btn1 = markup.Data(FRAMES[i], name, name)
		}
		if btn2 = markup.Data(" ", fmt.Sprintf("dull_%d", i)); i < len(NETS) {
			name := fmt.Sprintf("net_%d", i)
			btn2 = markup.Data(NETS[i], name, name)
		}
		bot.Handle(&btn1, f)
		bot.Handle(&btn2, f)
		rows = append(rows, markup.Row(btn1, btn2))
	}
	markup.Inline(rows...)

	return markup
}

func CreateOrderMarkup(bot *tele.Bot, f tele.HandlerFunc, can_add bool) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	if can_add {
		btn0 := markup.Data("Добавить", "order_0", "order_add")
		bot.Handle(&btn0, f)
		rows = append(rows, markup.Row(btn0))
	}
	btn1 := markup.Data("Закакать замер", "order_1", "order_measure")
	btn2 := markup.Data("Самовывоз", "order_2", "order_pickup")
	bot.Handle(&btn1, f)
	bot.Handle(&btn2, f)
	rows = append(rows, markup.Row(btn1, btn2))
	markup.Inline(rows...)

	return markup
}

func CreateButtonRows(bot *tele.Bot, f tele.HandlerFunc, options []string, prefix string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	for i, option := range options {
		btn := markup.Data(
			option,
			fmt.Sprintf("%s_%d", prefix, i),
			fmt.Sprintf("%s_%d", prefix, i))
		bot.Handle(&btn, f)
		rows = append(rows, markup.Row(btn))
	}
	markup.Inline(rows...)

	return markup
}
func CreateOptionCols(bot *tele.Bot, f tele.HandlerFunc, options []string, prefix string) *tele.ReplyMarkup {
	markup := &tele.ReplyMarkup{}
	var rows []tele.Row
	var btns []tele.Btn
	for i, option := range options {
		btn := markup.Data(
			option,
			fmt.Sprintf("%s_%d", prefix, i),
			fmt.Sprintf("%s_%d", prefix, i))
		bot.Handle(&btn, f)
		btns = append(btns, btn)
	}
	if len(btns)%2 > 0 {
		btns = append(btns, markup.Data(" ", "dull"))
	}
	for i := 0; i < len(btns); i += 2 {
		rows = append(rows, markup.Row(btns[i], btns[i+1]))
	}
	markup.Inline(rows...)

	return markup
}

func RequestMedia(ctx tele.Context) string {
	// ВЫЗВАТЬ ОПЕРАТОРА
	return "Я позвал оператора. Скоро он Вам ответит."
}
