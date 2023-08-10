package main

import "unsafe"

const PAGE_SIZE = 4096
const TABLE_MAX_PAGES = 100
const ROWS_PER_PAGE = (uint32)(PAGE_SIZE / ROW_SIZE)
const TABLE_MAX_ROWS = ROWS_PER_PAGE * PAGE_SIZE

type Table struct {
	numRows uint32
	pages   [TABLE_MAX_PAGES]unsafe.Pointer
}

// テーブルの行番号から、その位置を取得する
func rowSlot(table *Table, rowNum uint32) unsafe.Pointer {
	pageNum := rowNum / ROWS_PER_PAGE
	page := table.pages[pageNum]

	rowOffset := rowNum % ROWS_PER_PAGE
	byteOffset := rowOffset * uint32(ROW_SIZE)

	return unsafe.Pointer(uintptr(page) + uintptr(byteOffset))
}
