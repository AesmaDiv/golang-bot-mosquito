package sugar

import (
	"fmt"
	"strconv"
)

/* -------------- Type casting -------------- */

func ToString(x any) string {
	return Iif(x == nil, "", fmt.Sprintf("%v", x))
}
func ToInt(x any) int {
	return StrToInt(ToString(x))
}
func ToInt64(x any) int64 {
	return StrToInt64(ToString(x))
}
func ToUInt(x any) uint {
	return StrToUInt(ToString(x))
}
func ToUInt64(x any) uint64 {
	return StrToUInt64(ToString(x))
}
func ToFloat(x any) float32 {
	return StrToFloat(ToString(x))
}
func ToFloat64(x any) float64 {
	return StrToFloat64(ToString(x))
}
func ToBool(x any) bool {
	return StrToBool(ToString(x))
}
func StrToInt(x string) int {
	return int(StrToInt64(x))
}
func StrToInt64(x string) int64 {
	if val, err := strconv.ParseInt(x, 10, 64); err == nil {
		return val
	}
	return 0
}
func StrToUInt(x string) uint {
	return uint(StrToUInt64(x))
}
func StrToUInt64(x string) uint64 {
	if val, err := strconv.ParseUint(x, 10, 64); err == nil {
		return val
	}
	return 0
}
func StrToFloat(x string) float32 {
	return float32(StrToFloat64(x))
}
func StrToFloat64(x string) float64 {
	if val, err := strconv.ParseFloat(x, 64); err == nil {
		return val
	}
	return 0.0
}
func StrToBool(x string) bool {
	if val, err := strconv.ParseBool(x); err == nil {
		return val
	}
	return false
}
func AnyToBool(x any) bool {
	if val, ok := x.(bool); ok {
		return val
	}
	return false
}
func AnyToInt(x any) int {
	if val, ok := x.(int); ok {
		return val
	}
	return 0
}
func AnyToUInt(x any) uint {
	if val, ok := x.(uint); ok {
		return val
	}
	return 0
}
func AnyToInt64(x any) int64 {
	if val, ok := x.(int64); ok {
		return val
	}
	return 0
}
func AnyToUInt64(x any) uint64 {
	if val, ok := x.(uint64); ok {
		return val
	}
	return 0
}
func AnyToFloat(x any) float32 {
	if val, ok := x.(float32); ok {
		return val
	}
	return 0
}
func AnyToFloat64(x any) float64 {
	if val, ok := x.(float64); ok {
		return val
	}
	return 0
}
