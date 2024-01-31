package pd

func (df *SyncDataFrame) updateHeadIndexMap() {
	df.headLock.Lock()
	defer df.headLock.Unlock()
	df.DataFrame.updateHeadIndexMap()
}

func (df *SyncDataFrame) SetHeads(heads []string) {
	df.headLock.Lock()
	defer df.headLock.Unlock()
	df.DataFrame.SetHeads(heads)
}

func (df *SyncDataFrame) GetHeads() []string {
	df.headLock.RLock()
	defer df.headLock.RUnlock()
	return df.DataFrame.GetHeads()
}

func (df *SyncDataFrame) SetRows(rows [][]string) {
	df.rowLock.Lock()
	defer df.rowLock.Unlock()
	df.DataFrame.SetRows(rows)
}

func (df *SyncDataFrame) GetRows() [][]string {
	df.rowLock.RLock()
	defer df.rowLock.RUnlock()
	return df.DataFrame.GetRows()
}

func (df *SyncDataFrame) GetSheetName() string {
	df.headLock.RLock()
	defer df.headLock.RUnlock()
	return df.DataFrame.GetSheetName()
}

func (df *SyncDataFrame) SetSheetName(sheetName string) {
	df.headLock.Lock()
	defer df.headLock.Unlock()
	df.DataFrame.SetSheetName(sheetName)
}

func (df *SyncDataFrame) GetValue(rowIndex int, head any) (string, error) {
	df.rowLock.RLock()
	defer df.rowLock.RUnlock()
	return df.DataFrame.GetValue(rowIndex, head)
}

func (df *SyncDataFrame) SetValue(rowIndex int, head any, value string) error {
	df.rowLock.Lock()
	defer df.rowLock.Unlock()
	return df.DataFrame.SetValue(rowIndex, head, value)
}

func (df *SyncDataFrame) GetLength() int {
	df.rowLock.RLock()
	defer df.rowLock.RUnlock()
	return df.DataFrame.GetLength()
}

func (df *SyncDataFrame) ReadExcel(src string) error {
	df.rowLock.Lock()
	defer df.rowLock.Unlock()
	return df.DataFrame.ReadExcel(src)
}

func (df *SyncDataFrame) SaveExcel(dst string) error {
	df.rowLock.RLock()
	defer df.rowLock.RUnlock()
	return df.DataFrame.SaveExcel(dst)
}
