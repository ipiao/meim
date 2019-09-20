package util

import "strconv"

func IsStringIn(i string, arr []string) bool {
	for _, a := range arr {
		if i == a {
			return true
		}
	}
	return false
}

func StringsToInts(arr []string) []int {
	var is = make([]int, 0, len(arr))
	for _, a := range arr {
		i, _ := strconv.Atoi(a)
		is = append(is, i)
	}
	return is
}
