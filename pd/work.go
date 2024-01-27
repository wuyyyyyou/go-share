package pd

import (
	"fmt"
	"strings"

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

func (df *DataFrame) GetSheetName() string {
	if df.sheetName == nil {
		panic("sheet name is not set")
	}

	return *df.sheetName
}

func (df *DataFrame) SetSheetName(sheetName string) {
	df.sheetName = &sheetName
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

func (df *DataFrame) GetLength() int {
	return len(df.rows)
}

func (df *DataFrame) UniqueRows() {
	df.rows = lo.UniqBy(df.rows, func(slice []string) string {
		return strings.Join(slice, "\x1F")
	})
}

func (df *DataFrame) ReadExcel(src string) error {
	file, err := excelize.OpenFile(src)
	if err != nil {
		return err
	}
	defer ioutils.CloseQuietly(file)

	if df.sheetName == nil {
		df.SetSheetName(file.GetSheetList()[0])
	}
	rows, err := file.GetRows(df.GetSheetName())
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		return fmt.Errorf("sheet %s is empty", *df.sheetName)
	}

	df.SetHeads(rows[0])
	df.SetRows(rows[1:])
	df.updateHeadIndexMap()
	return nil
}

func (df *DataFrame) SaveExcel(dst string) error {
	file := excelize.NewFile()
	index, err := file.NewSheet(df.GetSheetName())
	if err != nil {
		return err
	}
	file.SetActiveSheet(index)

	for i, head := range df.GetHeads() {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)

		err = file.SetCellValue(df.GetSheetName(), cell, head)
		if err != nil {
			return err
		}
	}

	for i, row := range df.GetRows() {
		for j, cellValue := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)

			err = file.SetCellValue(df.GetSheetName(), cell, cellValue)
			if err != nil {
				return err
			}
		}
	}

	return file.SaveAs(dst)
}
