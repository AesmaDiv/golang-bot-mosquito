package telebot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	ss "golang-bot/sugar"
)

type TOrder struct {
	ID         int    //`db: "id"`
	CustomerID int64  //`db: "id_customer"`
	WorkerID   int64  //`db: "id_worker"`
	Frame      int    //`db: "frame"`
	Net        int    //`db: "net"`
	Sizes      []int  //`db: "sizes"`
	IsPickup   bool   //`db: "is_pickup "`
	DateTime   string //`db: "datetime "`
	Assigned   string //`db: "assigned "`
	Finished   string //`db: "finished "`
}

func (o TOrder) New() *TOrder {
	return &TOrder{
		ID:         0,
		CustomerID: 0,
		WorkerID:   0,
		Frame:      -1,
		Net:        -1,
		Sizes:      nil,
		IsPickup:   false,
		DateTime:   "",
		Assigned:   "",
		Finished:   "",
	}
}

func (o *TOrder) AddToDb(tele_id int64) int {
	sql := fmt.Sprintf(
		"Select * from updateorder(%d, %d, %d, '%s', %s)",
		o.CustomerID,
		o.Frame,
		o.Net,
		strings.Trim(fmt.Sprint(o.Sizes), "[]"),
		ss.ToString(o.IsPickup),
	)
	result := helper.Query(sql, []string{"id"})
	if len(result) == 0 {
		ss.Log("ERROR", "Order.AddToDB", "Встроенная функция UpdateOrder вернула ошибку")
		return 0
	}
	if val, ok := result[0]["id"]; ok {
		o.ID = ss.ToInt(val)
	}
	return o.ID
}

func (o TOrder) FromDb(id int) *TOrder {
	items := helper.Select("orders", []string{"*"}, map[string]any{"id": id})
	if len(items) > 0 {
		return o.FromMap(items[0])
	}
	return nil
}

func (o TOrder) GetOrdersFull(where string) []map[string]any {
	sql := fmt.Sprintf(`
		SELECT o.id,o.frame,o.net,o.datetime,o.sizes,o.is_pickup,o.id_worker,u.phone,u.fname From orders AS o
		JOIN users AS u ON o.id_customer=u.id_tele %s;`,
		where,
	)
	cols := []string{
		"id", "frame", "net", "datetime", "sizes", "is_pickup", "id_worker", "phone", "fname",
	}

	return helper.Query(sql, cols)
}

func (o TOrder) FromUser(user *TUser) *TOrder {
	if user.Order == nil {
		user.Order = TOrder{}.New()
	}
	return user.Order
}

func (o TOrder) FromMap(items map[string]any) *TOrder {
	sizes := []int{}
	if s := strings.Split(ss.ToString(items["sizes"]), " "); len(s) > 1 {
		sizes = ss.ArrayStr2Int(s)
	}
	return &TOrder{
		ID:         ss.ToInt(items["id"]),
		CustomerID: ss.ToInt64(items["id_customer"]),
		WorkerID:   ss.ToInt64(items["id_worker"]),
		Frame:      ss.ToInt(items["frame"]),
		Net:        ss.ToInt(items["net"]),
		Sizes:      sizes,
		IsPickup:   ss.ToBool(items["is_pickup"]),
		DateTime:   ss.ToString(items["datetime"]),
		Assigned:   ss.ToString(items["assigned"]),
		Finished:   ss.ToString(items["finished"]),
	}
}

func (o *TOrder) ParseOptions(text string) {
	data := strings.Split(text, "_")
	if len(data) == 2 {
		switch data[0] {
		case "frame":
			o.Frame = ss.StrToInt(data[1])
		case "net":
			o.Net = ss.StrToInt(data[1])
		}
	}
}

func (o *TOrder) ParseSizes(text string) {
	expr := regexp.MustCompile(`\d+`)
	vals := expr.FindAllString(text, -1)
	if len(vals) > 1 {
		l := len(vals) >> 1 << 1
		o.Sizes = make([]int, l)
		for i := 0; i < l; i++ {
			if v, e := strconv.Atoi(vals[i]); e == nil {
				o.Sizes[i] = v
			}
		}
	}
}

func (o *TOrder) Display(for_admin bool) string {
	prefix := ss.Iif(for_admin, "", "Вы выбрали:\n")
	suffix := ss.Iif(for_admin, "", fmt.Sprintf("\n%s\n", MSG_SIZES))
	per_sq_price, per_sq_text := o.getPriceForSquare()
	if per_sq_text == "" {
		return ""
	}
	per_size := o.getPriceForSize(per_sq_price)
	result := fmt.Sprintf("%s<code>%s%s</code>%s", prefix, per_sq_text, per_size, suffix)

	return result
}

func (o *TOrder) getPriceForSquare() (float32, string) {
	if o.Frame > -1 && o.Net > -1 {
		price := float32(PRICES[o.Frame][o.Net])
		result := fmt.Sprintf(
			" <i>рамка:</i>  %s\n"+
				" <i>сетка:</i>  %s\n"+
				" <i>цена:</i>   %.2f руб. за м²\n",
			FRAMES[o.Frame], NETS[o.Net], price)

		return price, result
	}

	return 0.0, ""
}

func (o *TOrder) getPriceForSize(price float32) string {
	var result string
	if len(o.Sizes) > 0 {
		result = " <i>размеры:</i>\n"
		for i := 0; i < len(o.Sizes); i += 2 {
			w := o.Sizes[i]
			h := o.Sizes[i+1]
			s := float32(w) * float32(h) / 1000000
			p := price * s
			result += fmt.Sprintf("   %dx%d = %.2f руб\n", w, h, p)
		}
	}

	return result
}

func (o *TOrder) UpdateWorker(id_order int, id_worker int64) int {
	o.WorkerID = id_worker
	o.Assigned = ss.Iif(id_worker > 0, ss.GetDateTime(), "")
	set := map[string]any{"id_worker": o.WorkerID, "assigned": o.Assigned}

	return helper.Update("orders", set, map[string]any{"id": id_order})
}
