package pd

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"

	"github.com/wuyyyyyou/go-share/ioutils"
	"github.com/wuyyyyyou/go-share/share"
)

func (df *DataFrame) updateHeadIndexMap() {
	df.headIndexMap = make(map[string]int)
	for i, head := range df.heads {
		df.headIndexMap[head] = i
	}
}

func (df *DataFrame) SetHeads(heads []string) {
	df.heads = heads
	df.updateHeadIndexMap()
}

func (df *DataFrame) GetHeads() []string {
	return df.heads
}

func (df *DataFrame) SetRows(rows [][]string) {
	df.rows = rows
}

func (df *DataFrame) GetRows() [][]string {
	return df.rows
}

// GetValue 返回索引处的值，如果索引超出范围，则返回错误，接受head的类型为string或int
func (df *DataFrame) GetValue(rowIndex int, head any) (string, error) {
	switch head := head.(type) {
	case string:
		index, ok := df.headIndexMap[head]
		if !ok {
			return "", fmt.Errorf("cannot find head %s", head)
		}
		if rowIndex >= len(df.rows) {
			return "", fmt.Errorf("row index %d out of range", rowIndex)
		}
		if index >= len(df.rows[rowIndex]) {
			return "", fmt.Errorf("head index %d out of range", index)
		}
		return df.rows[rowIndex][index], nil

	case int:
		if rowIndex >= len(df.rows) {
			return "", fmt.Errorf("row index %d out of range", rowIndex)
		}
		if head >= len(df.rows[rowIndex]) {
			return "", fmt.Errorf("head index %d out of range", head)
		}
		return df.rows[rowIndex][head], nil

	default:
		return "", fmt.Errorf("head type %T not supported", head)
	}
}

// SetValue 设置索引处的值，如果索引超出范围，会创建一个新的足够长的切片，接受head的类型为string或int
// 如果head为string类型，且不存在，则会自动添加到heads中
func (df *DataFrame) SetValue(rowIndex int, head any, value string) error {
	switch head := head.(type) {
	case string:
		index, ok := df.headIndexMap[head]
		if !ok {
			df.heads = append(df.heads, head)
			df.updateHeadIndexMap()
			index = df.headIndexMap[head]
		}
		if rowIndex >= len(df.rows) {
			df.rows = share.SetSliceValue(df.rows, rowIndex, make([]string, len(df.heads)))
		}
		df.rows[rowIndex] = share.SetSliceValue(df.rows[rowIndex], index, value)
		return nil

	case int:
		if rowIndex >= len(df.rows) {
			df.rows = share.SetSliceValue(df.rows, rowIndex, make([]string, len(df.heads)))
		}
		df.rows[rowIndex] = share.SetSliceValue(df.rows[rowIndex], head, value)
		return nil

	default:
		return fmt.Errorf("head type %T not supported", head)
	}
}

// AutoFillStruct sheet内容自动填充到结构体中，输入要求是一个结构体指针的切片的指针
func (df *DataFrame) AutoFillStruct(dest any) error {
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("outSlice must be a pointer to a slice")
	}

	elemType := destVal.Elem().Type().Elem()
	if elemType.Kind() != reflect.Ptr || elemType.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("outSlice must be a slice of pointer to struct")
	}

	for i := 0; i < df.GetLength(); i++ {
		newStructPtr := reflect.New(elemType.Elem())
		newStruct := newStructPtr.Elem()
		if err := df.fillStructFromSheet(i, newStruct, ""); err != nil {
			return err
		}
		destVal.Elem().Set(reflect.Append(destVal.Elem(), newStructPtr))
	}

	return nil
}

// fillStructFromSheet 递归处理嵌套结构体的字段
func (df *DataFrame) fillStructFromSheet(rowIndex int, val reflect.Value, prefix string) error {
	valType := val.Type()
	for j := 0; j < valType.NumField(); j++ {
		field := valType.Field(j)
		fieldVal := val.Field(j)

		if field.Anonymous {
			if err := df.fillStructFromSheet(rowIndex, fieldVal, prefix); err != nil {
				return err
			}
			continue
		}

		columnName := field.Tag.Get("pd")
		if columnName == "" {
			continue
		}

		// 处理带前缀的标签名
		if prefix != "" {
			columnName = prefix + "_" + columnName
		}

		if !fieldVal.CanSet() {
			return fmt.Errorf("cannot set field %s", field.Name)
		}

		switch fieldVal.Kind() {
		case reflect.String:
			value, _ := df.GetValue(rowIndex, columnName)
			fieldVal.SetString(value)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			value, _ := df.GetValue(rowIndex, columnName)
			intVal, _ := strconv.ParseInt(value, 10, 64)
			fieldVal.SetInt(intVal)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			value, _ := df.GetValue(rowIndex, columnName)
			uintVal, _ := strconv.ParseUint(value, 10, 64)
			fieldVal.SetUint(uintVal)
		case reflect.Float32, reflect.Float64:
			value, _ := df.GetValue(rowIndex, columnName)
			floatVal, _ := strconv.ParseFloat(value, 64)
			fieldVal.SetFloat(floatVal)
		case reflect.Bool:
			value, _ := df.GetValue(rowIndex, columnName)
			boolVal, _ := strconv.ParseBool(value)
			fieldVal.SetBool(boolVal)
		case reflect.Struct:
			if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
				value, _ := df.GetValue(rowIndex, columnName)
				timeVal, _ := time.Parse(time.RFC3339, value)
				fieldVal.Set(reflect.ValueOf(timeVal))
			} else {
				if err := df.fillStructFromSheet(rowIndex, fieldVal, columnName); err != nil {
					return err
				}
				continue
			}
		case reflect.Ptr:
			if fieldVal.Type().Elem().Kind() == reflect.Struct {
				if fieldVal.IsNil() {
					fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
				}
				if err := df.fillStructFromSheet(rowIndex, fieldVal.Elem(), columnName); err != nil {
					return err
				}
				continue
			}
		default:
			return fmt.Errorf("unsupported field type: %v", fieldVal.Kind())
		}
	}
	return nil
}

// AutoFillSheet 结构体内容填充到excel表格中，会覆盖原本的内容，输入要求是一个结构体指针切片
func (df *DataFrame) AutoFillSheet(dest any) error {
	df.SetRows([][]string{})
	df.SetHeads([]string{})
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Slice {
		return fmt.Errorf("inputSlice must be a slice")
	}

	elemType := destVal.Type().Elem()
	if elemType.Kind() != reflect.Ptr || elemType.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("inputSlice must be a slice of pointer to struct")
	}

	for i := 0; i < destVal.Len(); i++ {
		elemVal := destVal.Index(i).Elem()
		if err := df.fillStructFields(i, elemVal, ""); err != nil {
			return err
		}
	}

	return nil
}

// fillStructFields 递归处理嵌套结构体的字段
func (df *DataFrame) fillStructFields(rowIndex int, val reflect.Value, prefix string) error {
	valType := val.Type()
	for j := 0; j < valType.NumField(); j++ {
		field := valType.Field(j)
		fieldVal := val.Field(j)

		if field.Anonymous {
			if err := df.fillStructFields(rowIndex, fieldVal, prefix); err != nil {
				return err
			}
			continue
		}

		columnName := field.Tag.Get("pd")
		if columnName == "" {
			continue
		}

		// 处理带前缀的标签名
		if prefix != "" {
			columnName = prefix + "_" + columnName
		}

		var inputVal string
		switch fieldVal.Kind() {
		case reflect.String:
			inputVal = fieldVal.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			inputVal = strconv.FormatInt(fieldVal.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			inputVal = strconv.FormatUint(fieldVal.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			inputVal = strconv.FormatFloat(fieldVal.Float(), 'f', -1, 64)
		case reflect.Bool:
			inputVal = strconv.FormatBool(fieldVal.Bool())
		case reflect.Struct:
			if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
				inputVal = fieldVal.Interface().(time.Time).Format(time.RFC3339)
			} else {
				if err := df.fillStructFields(rowIndex, fieldVal, columnName); err != nil {
					return err
				}
				continue
			}
		case reflect.Ptr:
			if fieldVal.Type().Elem().Kind() == reflect.Struct {
				if fieldVal.IsNil() {
					fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
				}
				if err := df.fillStructFields(rowIndex, fieldVal.Elem(), columnName); err != nil {
					return err
				}
				continue
			}
		default:
			return fmt.Errorf("unsupported field type: %v", fieldVal.Kind())
		}

		if err := df.SetValue(rowIndex, columnName, inputVal); err != nil {
			return err
		}
	}
	return nil
}

func (df *DataFrame) GetLength() int {
	return len(df.rows)
}

func (df *DataFrame) UniqueRows() {
	df.rows = lo.UniqBy(df.rows, func(slice []string) string {
		return strings.Join(slice, "\x1F")
	})
}

func (e *Excel) AppendSheet(dfs ...*DataFrame) {
	for _, df := range dfs {
		sheetName := df.sheetName
		if sheetName == "" {
			sheetName = "Sheet1"
		}
		if !lo.Contains(e.SheetNames, sheetName) {
			e.SheetNames = append(e.SheetNames, sheetName)
		}
		e.DataFramesMap[sheetName] = df
	}
}

func (e *Excel) ReadExcelAllSheet(src string) error {
	file, err := excelize.OpenFile(src)
	if err != nil {
		return err
	}
	defer ioutils.CloseQuietly(file)

	e.SheetNames = file.GetSheetList()
	for _, sheetName := range e.SheetNames {
		df := NewDataFrame(sheetName)
		rows, err := file.GetRows(sheetName)
		if err != nil {
			return err
		}
		if len(rows) > 0 {
			df.SetHeads(rows[0])
			df.SetRows(rows[1:])
			df.updateHeadIndexMap()
		}
		e.DataFramesMap[sheetName] = df
	}

	return nil
}

func (e *Excel) SaveExcelAllSheet(dst string) error {
	file := excelize.NewFile()

	for sheetName, df := range e.DataFramesMap {
		index, err := file.NewSheet(sheetName)
		if err != nil {
			return err
		}
		file.SetActiveSheet(index)

		for i, head := range df.GetHeads() {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)

			err = file.SetCellValue(sheetName, cell, head)
			if err != nil {
				return err
			}
		}

		for i, row := range df.GetRows() {
			for j, cellValue := range row {
				cell, _ := excelize.CoordinatesToCellName(j+1, i+2)

				err = file.SetCellValue(sheetName, cell, cellValue)
				if err != nil {
					return err
				}
			}
		}
	}

	if !lo.Contains(e.SheetNames, "Sheet1") {
		err := file.DeleteSheet("Sheet1")
		if err != nil {
			return err
		}
	}

	return file.SaveAs(dst)
}

func (df *DataFrame) ReadCsv(src string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer ioutils.CloseQuietly(file)

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return fmt.Errorf("csv file is empty")
	}

	df.SetHeads(records[0])
	df.SetRows(records[1:])
	return nil
}

func (df *DataFrame) SaveCsv(dst string) error {
	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer ioutils.CloseQuietly(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(df.heads); err != nil {
		return err
	}

	if err := writer.WriteAll(df.rows); err != nil {
		return err
	}

	return nil
}
