package utils

import "strconv"

func StrToFloat32(val string) float32 {
	f, _ := strconv.ParseFloat(val, 32)
	return float32(f)
}

func StrToInt32(val string) int32 {
	i, _ := strconv.Atoi(val)
	return int32(i)
}
