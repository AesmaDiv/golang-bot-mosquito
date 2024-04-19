package sugar

func ArrayStr2Int(arr []string) []int {
	result := make([]int, len(arr))
	for i, s := range arr {
		result[i] = StrToInt(s)
	}
	return result
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
