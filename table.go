package main

import "fmt"

const PAGE_SIZE = 4096
const TABLE_MAX_PAGES = 100
const ROWS_PER_PAGE = PAGE_SIZE / ROW_SIZE
const TABLE_MAX_ROWS = ROWS_PER_PAGE * TABLE_MAX_PAGES

type Page [PAGE_SIZE]byte

type Table struct {
	numRows uint32
	pages   [TABLE_MAX_PAGES]Page
}

// テーブルの行番号から、その行が保存されているスライスを取得する
func rowSlot(table *Table, rowNumUint32 uint32) []byte {
	rowNum := int(rowNumUint32)
	pageNum := int(rowNum) / ROWS_PER_PAGE
	page := table.pages[pageNum]

	rowOffset := rowNum % ROWS_PER_PAGE
	byteOffset := rowOffset * ROW_SIZE

	fmt.Println(rowNum, pageNum)

	row := page[byteOffset : byteOffset+ROW_SIZE]

	fmt.Println(row)

	return row
}
