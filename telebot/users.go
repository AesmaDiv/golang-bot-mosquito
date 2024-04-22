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
	IDChat    int64
	IDTele    int64
	UserName  string
	FirstName string
	LastName  string
	Phone     string
	Visited   string
	IsAdmin   bool
	IsBanned  bool
	// поля в кэше
	Order        *TOrder
	Status       string
	MessageOrder *tele.Message
	MessageLast  *tele.Message
}

var _users = make(map[int64]*TUser)

func (u TUser) New(id int64, names ...string) *TUser {
	var uname, fname, lname string
	if len(names) > 0 {
		uname = names[0]
	}
	if len(names) > 1 {
		fname = names[1]
	}
	if len(names) > 2 {
		lname = names[2]
	}
	return &TUser{
		IDTele:    id,
		IDChat:    id,
		UserName:  uname,
		FirstName: fname,
		LastName:  lname,
		Phone:     "",
		Visited:   ss.GetDateTime(),
		IsAdmin:   false,
		IsBanned:  false,

		Order:        nil,
		Status:       EXP_START,
		MessageOrder: nil,
		MessageLast:  nil,
	}
}

func (u TUser) FromMap(row map[string]any) *TUser {
	return &TUser{
		IDTele:    ss.AnyToInt64(row["id_tele"]),
		IDChat:    ss.AnyToInt64(row["id_chat"]),
		UserName:  ss.ToString(row["uname"]),
		FirstName: ss.ToString(row["fname"]),
		LastName:  ss.ToString(row["lname"]),
		Phone:     ss.ToString(row["phone"]),
		Visited:   ss.ToString(row["visit"]),
		IsAdmin:   ss.AnyToBool(row["is_admin"]),
		IsBanned:  ss.AnyToBool(row["is_banned"]),

		Order:        nil,
		Status:       EXP_START,
		MessageOrder: nil,
		MessageLast:  nil,
	}
}
func (u TUser) Get(id int64, names ...string) *TUser {
	ss.Log("INFO", "user.Get", fmt.Sprintf("Идентификация пользователя %d", id))
	var user *TUser
	// сначала ищу в кэше
	user, ok := _users[id]
	if !ok {
		ss.Log("INFO", "user.Get", fmt.Sprintf("Пользователь %d НЕ найден в кэше", id))
		// если не найден, ищу в БД
		user = TUser{}.FromDb(id)
		if user == nil {
			ss.Log("INFO", "user.Get", fmt.Sprintf("Пользователь %d НЕ найден в БД. Создали нового", id))
			// если нет в БД, то создаю нового
			user = TUser{}.New(id, names...)
			user.AddToDb()
		} else {
			ss.Log("INFO", "user.Get", fmt.Sprintf("Пользователь %d найден в БД", id))
		}
		user.AddToCache()
	} else {
		ss.Log("INFO", "user.Get", fmt.Sprintf("Пользователь %d найден в кэше", id))
	}

	return user
}

func (u TUser) FromDb(id int64) *TUser {
	user := helper.Select("users", []string{"*"}, map[string]any{"id_tele": id})
	if len(user) == 0 {
		ss.Log("INFO", "user.fromDB",
			fmt.Sprintf("Пользователь %d НЕ найден в БД", id))
		return nil
	} else {
		ss.Log("INFO", "user.fromDB",
			fmt.Sprintf("Пользователь %d найден в БД", id))
	}

	return TUser{}.FromMap(user[0])
}

func (u TUser) AddToDb() {
	go helper.Insert(
		"users",
		[]map[string]any{{
			"id_tele": u.IDTele,
			"id_chat": u.IDChat,
			"uname":   u.UserName,
			"fname":   u.FirstName,
		}},
	)
}

func (u TUser) AddToCache() {
	ss.Log("INFO", "user.addToCache",
		fmt.Sprintf("Добавление пользователя %d в кэш", u.IDTele))
	_users[u.IDTele] = &u
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
		map[string]any{"id_tele": u.IDTele},
	)
}

func (u TUser) UpdateVisit() {
	u.Visited = ss.GetDateTime()
	res := helper.Update(
		"users",
		map[string]any{"visit": u.Visited},
		map[string]any{"id_tele": u.IDTele},
	)
	if res > 0 {
		ss.Log("INFO", "UpdateUserVisit", fmt.Sprintf("Обновлено время посещения для пользователя %d", u.IDTele))
	} else {
		ss.Log("FAILED", "UpdateUserVisit", fmt.Sprintf("Не удалось обновить время посещения для пользователя %d", u.IDTele))
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
