package pd

type Excel struct {
	SheetNames    []string
	DataFramesMap map[string]*DataFrame
}

func NewExcel() *Excel {
	return &Excel{
		SheetNames:    []string{},
		DataFramesMap: map[string]*DataFrame{},
	}
}

type DataFrame struct {
	sheetName    string
	heads        []string
	rows         [][]string
	headIndexMap map[string]int
}

func NewDataFrame(sheetName string) *DataFrame {
	return &DataFrame{
		heads:        []string{},
		rows:         [][]string{},
		sheetName:    sheetName,
		headIndexMap: make(map[string]int),
	}
}
