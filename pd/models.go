package pd

import (
	"sync"
)

type DataFrame struct {
	sheetName    *string
	heads        []string
	rows         [][]string
	headIndexMap map[string]int
}

func NewDataFrame(Args ...string) *DataFrame {
	if len(Args) == 0 || Args[0] == "" {
		return &DataFrame{
			heads:        []string{},
			rows:         [][]string{},
			headIndexMap: make(map[string]int),
		}
	}

	return &DataFrame{
		heads:        []string{},
		rows:         [][]string{},
		sheetName:    &Args[0],
		headIndexMap: make(map[string]int),
	}
}

// 线程安全的df
type SyncDataFrame struct {
	*DataFrame
	headLock sync.RWMutex
	rowLock  sync.RWMutex
}

func NewSyncDataFrame(Args ...string) *SyncDataFrame {
	if len(Args) == 0 || Args[0] == "" {
		return &SyncDataFrame{
			DataFrame: &DataFrame{
				heads:        []string{},
				rows:         [][]string{},
				headIndexMap: make(map[string]int),
			},
		}
	}

	return &SyncDataFrame{
		DataFrame: &DataFrame{
			heads:        []string{},
			rows:         [][]string{},
			sheetName:    &Args[0],
			headIndexMap: make(map[string]int),
		},
	}
}
