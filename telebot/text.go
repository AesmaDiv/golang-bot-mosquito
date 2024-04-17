package telebot

const (
	MSG_START    = "Чем я могу быть Вам полезен?"
	MSG_FRAME    = "Выберите <b>РАМКУ</b>"
	MSG_NET      = "Выберите <b>СЕТКУ</b>"
	MSG_SIZES    = "<i>Вы можете указать несколько размеров рамок в мм, в формате <b>Ш*В</b>, разделяя пробелом</i>"
	MSG_ASKPHONE = "Для подтверждения заказа укажите, пожалуйста, номер Вашего телефона и Имя."
	MSG_WILLCALL = "С Вами свяжутся в течении 30 минут"
	MSG_MEDIA    = "Прикрепите фото/видеофайл к сообщению" // содержащему <b>Ваш номер телефона и Имя</b>"
	MSG_ERRPHONE = "Прошу прощения. Я не смог разобрать Ваш номер телефона."
	MSG_ADMIN    = "Приветствую тебя, Повелитель!"
	MSG_NOSTART  = "Прошу прощения, но меня можно стартовать только в приватном чате"

	ADMIN_CUSTOMERS   = "admin_0"
	ADMIN_MY_ORDERS   = "admin_1"
	ADMIN_FREE_ORDERS = "admin_2"

	BTN_SHOW_OPTIONS = "option_0"
	BTN_SEND_MEDIA   = "option_1"
	BTN_REQUEST_CALL = "option_2"
	BTN_ADMIN        = "option_3"

	EXP_CONTACT = "expectContact"
	EXP_SIZES   = "expectSizes"
	EXP_MEDIA   = "expectMedia"

	ORDER_NEW   = "<b>≡≡==----- НОВЫЙ ЗАКАЗ -----==≡≡</b>"
	ORDER_FREE  = "<b>≡≡==--- СВОБОДНЫЙ ЗАКАЗ ---==≡≡</b>"
	ORDER_YOUR  = "<b>≡≡==------ ВАШ ЗАКАЗ ------==≡≡</b>"
	ORDER_MEDIA = "<b>≡≡==--- ОЦЕНКА  РЕМОНТА ---==≡≡</b>"

	ERR_RESTART = "Ой! Что то я запутался. Пожалуйства, перезагрузите меня, нажав 'старт' ещё раз."
)

var (
	ON_START = []string{
		"Калькулятор москитных сеток",
		"Оценка ремонта окон по медиафайлу",
		"Заказать звонок",
	}
	ON_ORDER = []string{
		"Заказать замер",
		"Самовывоз",
	}
	ON_ADMIN = []string{
		"Клиенты",
		"Мои заказы",
		"Свободные заказы",
	}
	FRAMES = []string{
		"Рамочная 25мм",
		"Раздвижная",
		"Рулонная",
		"Вставная изнутри",
		"Дверная",
		"Плиссе",
	}
	NETS = []string{
		"Стандарт",
		"Антикошка",
		"Антипыль",
		"Антипыльца 95%",
		"Алюминиевая",
		"Нержавейка",
	}
	PRICES = [][]int{
		{640, 1700, 1350, 1100, 2100, 1500, 2700, 3050, 2950, 3200, 5000},
		{1400, 2550, 2200, 1950, 2950, 2350, 3550, 3900, 3800, 4050, 6000},
		{1050, 2200, 1850, 1600, 2600, 2000, 3200, 3550, 3450, 3750, 6000},
		{950, 2100, 1750, 1500, 2500, 1900, 3100, 3450, 3350, 3650, 5800},
		{1950, 3100, 2750, 2500, 3500, 2900, 4100, 4450, 4350, 4650, 7200},
		{1750, 2900, 2550, 2300, 3300, 2700, 3900, 4250, 4150, 4450, 7200},
		{1700, 2850, 2500, 2250, 3250, 2650, 3850, 4200, 4100, 4400, 6300},
		{2300, 3450, 3100, 2850, 3850, 3250, 4450, 4800, 4700, 5000, 7300},
		{2900, 4050, 3700, 3450, 4450, 3850, 5050, 5400, 5300, 5600, 8300},
		{2500, 3500, 3200, 2900, 3950, 3350, 4550, 4900, 4800, 5100, 7900},
	}
)
