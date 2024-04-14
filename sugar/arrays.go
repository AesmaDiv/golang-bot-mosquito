package sugar

func ArrayStr2Int(arr []string) []int {
	result := make([]int, len(arr))
	for i, s := range arr {
		result[i] = ToInt(s)
	}
	return result
}
