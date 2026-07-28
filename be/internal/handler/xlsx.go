package handler

import "github.com/xuri/excelize/v2"

// openXlsx opens an xlsx file for reading (separate helper so handlers stay lean).
func openXlsx(path string) (*excelize.File, error) {
	return excelize.OpenFile(path)
}
