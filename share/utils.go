package share

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime/debug"
	"strings"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
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

func Except() {
	if err := recover(); err != nil {
		fmt.Printf("Recovered from panic: %v\n", err)
		fmt.Printf("Stack trace:\n%s\n", debug.Stack())
	}
}

func EnsureHttpPrefix(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "http://" + url
	}
	return url
}

// ConvertEncoding 用于检测和转换编码的函数
func ConvertEncoding(contentType string, body []byte) ([]byte, error) {
	var e encoding.Encoding
	var name string

	// 检测内容编码
	_, name, _ = charset.DetermineEncoding(body, contentType)

	// 根据不同的编码进行处理
	switch name {
	case "gbk", "gb18030":
		e = simplifiedchinese.GB18030
	case "big5":
		e = traditionalchinese.Big5
	case "windows-1252":
		e = charmap.Windows1252
	// 这里可以添加更多编码的处理
	default:
		e = nil
	}

	if e != nil {
		// 转换编码
		return io.ReadAll(transform.NewReader(bytes.NewReader(body), e.NewDecoder()))
	}

	// 如果没有特殊编码处理，则原样返回
	return body, nil
}

// SHA256 SHA256哈希，返回32位字节切片
func SHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// SHA256ToString SHA256哈希，返回64为长度字符串
func SHA256ToString(data []byte) string {
	return hex.EncodeToString(SHA256(data))
}

// MapToStruct 将 map 转换为指定的结构体
func MapToStruct(m map[string]any, result any) error {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, result)
}

// StructToMap 将结构体转换为 map
func StructToMap(obj any) (map[string]any, error) {
	result := make(map[string]any)
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// TrimStructStrings 接受一个结构体指针，对所有 string 字段进行 TrimSpace 操作
func TrimStructStrings(s any) error {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("the input is not a pointer")
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("the input is not a struct pointer")
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		// 如果字段是字符串类型，并且可以修改
		if field.Kind() == reflect.String && field.CanSet() {
			trimmed := strings.TrimSpace(field.String())
			field.SetString(trimmed)
		}
	}

	return nil
}
