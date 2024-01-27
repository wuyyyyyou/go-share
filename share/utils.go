package share

import (
	"os"
)

// SetSliceValue 设置切片的值，如果索引超出范围，则创建一个新的足够长的切片
func SetSliceValue[T any](s []T, index int, value T) []T {
	// 如果索引在范围内，则直接设置值
	if index < len(s) {
		s[index] = value
		return s
	}

	// 如果索引超出范围，则创建一个新的足够长的切片
	newSlice := make([]T, index+1)
	copy(newSlice, s)
	newSlice[index] = value
	return newSlice
}

func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
