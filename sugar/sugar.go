package sugar

import (
	"fmt"
	"log"
	"time"
)

func GetDate() string {
	return time.Now().Format("2006-01-02")
}
func GetDateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func GenGreeting() string {
	switch hour := time.Now().Hour(); {
	case hour < 5:
		return "Доброй ночи"
	case hour < 10:
		return "Доброе утро"
	case hour < 16:
		return "Добрый день"
	default:
		return "Добрый вечер"
	}
}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Iif[T comparable](_if bool, _then T, _else T) T {
	if _if {
		return _then
	}
	return _else
}

func Log(state string, name, msg string) {
	fmt.Printf("%s [%s] %s:: %s\n", GetDateTime(), state, name, msg)
}

func RemoveFromArray[T comparable](arr []T, index int) []T {
	if index < 0 || index >= len(arr) {
		return arr
	}
	if index == len(arr)-1 {
		return arr[:index]
	}
	return append(arr[:index], arr[index+1:]...)
}

func ArrLastRef[T comparable](arr []T) *T {
	if len(arr) == 0 {
		return nil
	}
	return &arr[len(arr)-1]
}
func ArrLastVal(arr []any) any {
	if len(arr) == 0 {
		return nil
	}
	return &arr[len(arr)-1]
}
