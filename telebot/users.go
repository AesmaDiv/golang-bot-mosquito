package telebot

import (
	"fmt"
	db "golang-bot/database"
	ss "golang-bot/sugar"
	"regexp"
	"strings"

	tele "gopkg.in/telebot.v3"
)

type TUser struct {
	// поля из БД
	ChatID    int64
	TeleID    int64
	UserName  string
	FirstName string
	LastName  string
	Phone     string
	Address   string
	IsAdmin   bool
	Visited   string
	// поля в кэше
	Order        *TOrder
	Status       string
	MessageOrder *tele.Message
	MessageLast  *tele.Message
}

var USERS = make(map[int64]*TUser)

func (u TUser) New(id int64, chat_id int64, username string, firstname string) *TUser {
	return &TUser{
		TeleID:    id,
		ChatID:    chat_id,
		UserName:  username,
		FirstName: firstname,

		Visited:      ss.GetDateTime(),
		Order:        nil,
		Status:       "idle",
		MessageOrder: nil,
	}
}

func (u TUser) Get(id int64) *TUser {
	var user *TUser
	// сначала ищу в кэше
	user, found := USERS[id]
	if !found {
		// если не найден, ищу в БД
		user = TUser{}.FromDb(helper, id)
		// если нет в БД, то и ладно, потом создадим
	}

	return user
}

func (u TUser) FromDb(helper db.Helper, id int64) *TUser {
	user := helper.Select("users", []string{"*"}, map[string]any{"id_tele": id})
	if len(user) == 0 {
		return nil
	}

	return TUser{}.FromMap(user[0])
}

func (u TUser) FromMap(row map[string]any) *TUser {
	return &TUser{
		TeleID:    ss.ToInt64(row["id_tele"]),
		ChatID:    ss.ToInt64(row["id_chat"]),
		UserName:  ss.ToString(row["uname"]),
		FirstName: ss.ToString(row["fname"]),
		LastName:  ss.ToString(row["lname"]),
		Phone:     ss.ToString(row["phone"]),
		Address:   ss.ToString(row["address"]),
		IsAdmin:   ss.ToBool(row["is_admin"]),
		Visited:   ss.ToString(row["visit"]),

		Order:        nil,
		Status:       "idle",
		MessageOrder: nil,
	}
}

func (u TUser) AddToDb(helper db.Helper) {
	go helper.Insert(
		"users",
		[]map[string]any{{
			"id_tele": u.TeleID,
			"id_chat": u.ChatID,
			"uname":   u.UserName,
			"fname":   u.FirstName,
		}},
	)
}

func (u TUser) AddToCache() {
	USERS[u.TeleID] = &u
}

func (u TUser) Display() string {
	return fmt.Sprintf(
		"%s <b>%s %s</b>\n<i>%s</i>\nпоследний визит:\n%s\n----------------\n",
		u.Phone,
		u.FirstName,
		u.LastName,
		u.UserName,
		u.Visited,
	)
}

func (u TUser) DBUpdate_Contact(helper db.Helper) {
	user_data := map[string]any{"phone": u.Phone}
	// обновляем имя пользователя, только если оно было передано
	if u.FirstName != "" {
		user_data["fname"] = u.FirstName
	}
	// обновляем данные в БД
	go helper.Update(
		"users",
		user_data,
		map[string]any{"id_tele": u.TeleID},
	)
}

func (u TUser) DBUpdate_Visit(helper db.Helper) {
	u.Visited = ss.GetDateTime()
	res := helper.Update(
		"users",
		map[string]any{"visit": u.Visited},
		map[string]any{"id_tele": u.TeleID},
	)
	if res > 0 {
		ss.Log("SUCCESS", "UpdateUserVisit", u.UserName)
	} else {
		ss.Log("FAILED", "UpdateUserVisit", u.UserName)
	}
}

func (u *TUser) ParseContact(msg string) {
	find_phone := regexp.MustCompile(`(\d+)`)
	find_uname := regexp.MustCompile(`([a-zA-Zа-яА-Я]+)`)
	may_be_phone := find_phone.FindAllString(msg, -1)
	may_be_uname := find_uname.FindAllString(msg, -1)

	if uname := strings.Join(may_be_uname, " "); uname != "" {
		u.FirstName = uname
	}
	if phone := validatePhone(strings.Join(may_be_phone, "")); phone != "" {
		u.Phone = phone
	}
}

func validatePhone(phone string) string {
	if len(phone) < 10 {
		return ""
	}
	if strings.HasPrefix(phone, "8") {
		phone = strings.Replace(phone, "8", "", 1)
	}
	phone = strings.TrimPrefix(phone, "7")

	return fmt.Sprintf("+7%s", phone)
}
