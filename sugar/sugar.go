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

func GetMaxLength[T comparable](arr1 []T, arr2 []T) int {
	len1 := len(arr1)
	len2 := len(arr2)
	return Iif(len1 >= len2, len1, len2)
}

func GetMinLength[T comparable](arr1 []T, arr2 []T) int {
	len1 := len(arr1)
	len2 := len(arr2)
	return Iif(len1 <= len2, len1, len2)
}

func ReorderDesc[T comparable](arr1, arr2 []T) ([]T, []T, int, int) {
	len1 := len(arr1)
	len2 := len(arr2)
	if len1 >= len2 {
		return arr1, arr2, len1, len2
	}
	return arr2, arr1, len2, len1
}

func ToString(x any) string {
	return Iif(x == nil, "", fmt.Sprintf("%v", x))
}
func ToBool(x any) bool {
	if val, ok := x.(bool); ok {
		return val
	}
	return false
}
func ToInt(x any) int {
	if val, ok := x.(int); ok {
		return val
	}
	return 0
}
func ToUInt(x any) uint {
	if val, ok := x.(uint); ok {
		return val
	}
	return 0
}
func ToInt64(x any) int64 {
	if val, ok := x.(int64); ok {
		return val
	}
	return 0
}
func ToUInt64(x any) uint64 {
	if val, ok := x.(uint64); ok {
		return val
	}
	return 0
}
func ToFloat(x any) float32 {
	if val, ok := x.(float32); ok {
		return val
	}
	return 0
}
func ToFloat64(x any) float64 {
	if val, ok := x.(float64); ok {
		return val
	}
	return 0
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
