package pd

type DataFrame struct {
	sheetName    string
	heads        []string
	rows         [][]string
	headIndexMap map[string]int
}

func NewDataFrame(Args ...string) *DataFrame {
	if len(Args) == 0 || Args[0] == "" {
		return &DataFrame{
			heads:        []string{},
			rows:         [][]string{},
			sheetName:    "Sheet1",
			headIndexMap: make(map[string]int),
		}
	}

	return &DataFrame{
		heads:        []string{},
		rows:         [][]string{},
		sheetName:    Args[0],
		headIndexMap: make(map[string]int),
	}
}
