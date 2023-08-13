package main

import "unsafe"

const (
	PAGE_SIZE       = 4096
	TABLE_MAX_PAGES = 100
	ROWS_PER_PAGE   = PAGE_SIZE / ROW_SIZE
	TABLE_MAX_ROWS  = ROWS_PER_PAGE * TABLE_MAX_PAGES
)

const (
	ID_OFFSET       = 0
	USERNAME_OFFSET = ID_OFFSET + int(unsafe.Sizeof(int(0)))
	EMAIL_OFFSET    = USERNAME_OFFSET + int(unsafe.Sizeof([32]byte{}))
	ID_SIZE         = int(unsafe.Sizeof(int(0)))
	USERNAME_SIZE   = int(unsafe.Sizeof([32]byte{}))
	EMAIL_SIZE      = int(unsafe.Sizeof([256]byte{}))
	ROW_SIZE        = ID_SIZE + USERNAME_SIZE + EMAIL_SIZE
)

type Page [PAGE_SIZE]byte

type Table struct {
	numRows uint32
	pages   [TABLE_MAX_PAGES]Page
}

// メモリ上の行の位置
type RowSlot struct {
	pageNum  int
	rowStart int
	rowEnd   int
}

// 行番号から、行のスロットを取得する
func (table *Table) getRowSlot(rowNumUint32 uint32) RowSlot {
	rowNum := int(rowNumUint32)
	pageNum := int(rowNum) / ROWS_PER_PAGE
	rowOffset := rowNum % ROWS_PER_PAGE
	byteOffset := rowOffset * ROW_SIZE

	return RowSlot{pageNum: pageNum, rowStart: byteOffset, rowEnd: byteOffset + ROW_SIZE}
}

// 行を挿入する
func (table *Table) insertRow(row *Row) {
	rs := table.getRowSlot(table.numRows)
	sourceBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(row))

	// ページに書き込む
	copy(table.pages[rs.pageNum][rs.rowStart:rs.rowEnd], sourceBytes[:])
	table.numRows++
}

// 行を取得する
func (table *Table) getRowByRowNum(rowNum uint32) Row {
	rs := table.getRowSlot(rowNum)

	// ページから、Row構造体に書き込む
	row := &Row{}
	destinationBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(row))
	copy(destinationBytes[:], table.pages[rs.pageNum][rs.rowStart:rs.rowEnd])

	return *row
}
