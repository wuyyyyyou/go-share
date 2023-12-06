package share

import (
	"io"
	"log"
)

// SetSliceValue 设置切片的值，如果索引超出范围，则创建一个新的足够长的切片
func SetSliceValue[T any](s []T, index int, value T) []T {
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

// CloseAll 关闭所有的io.Closer
func CloseAll(cr ...io.Closer) {
	for _, c := range cr {
		if c != nil {
			if err := c.Close(); err != nil {
				log.Println(err)
			}
		}
	}
}
