package table

import (
	"toydb-go/persistence"
	"unsafe"
)

type Row struct {
	Id       int
	Username [32]byte
	Email    [256]byte
}

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

type Table struct {
	pager       persistence.Pager
	rootPageNum uint32
}

// func GetNumRows(table *Table) uint32 {
// 	return table.numRows
// }

// 行を挿入する
func (table *Table) InsertRow(row *Row) {
	cursor := tableEnd(table)
	persistence.GetNumCells(&table.pager, cursor.PageNum)
	leafNodeInsert(&cursor, uint32(row.Id), row)
}

// 行を取得する
func (table *Table) GetRowByRowNum(pageNum uint32, cellNum uint32) Row {
	rs := persistence.GetRowSlot(pageNum, cellNum)

	// ページから、Row構造体に書き込む
	row := &Row{}
	destinationBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(row))
	copy(destinationBytes[:], table.pager.GetRow(rs))

	return *row
}

type Cursor struct {
	table      *Table
	PageNum    uint32
	CellNum    uint32
	EndOfTable bool
}

func TableStart(table *Table) Cursor {
	endOfTable := false
	numCells := persistence.GetNumCells(&table.pager, table.rootPageNum)
	if numCells == 0 {
		endOfTable = true
	}

	return Cursor{
		table:      table,
		PageNum:    table.rootPageNum,
		CellNum:    0,
		EndOfTable: endOfTable,
	}
}

func tableEnd(table *Table) Cursor {
	numCells := persistence.GetNumCells(&table.pager, table.rootPageNum)

	return Cursor{
		table:      table,
		PageNum:    table.rootPageNum,
		CellNum:    numCells,
		EndOfTable: true,
	}
}

// カーソルのページ上での位置を返す
func cursorValue(cursor *Cursor) persistence.RowSlot {
	return persistence.GetRowSlot(cursor.PageNum, cursor.CellNum)
}

// カーソルを1つ進める
func CursorAdvance(cursor *Cursor) {
	numCells := persistence.GetNumCells(&cursor.table.pager, cursor.PageNum)
	cursor.CellNum += 1

	if cursor.CellNum >= numCells {
		cursor.EndOfTable = true
	}
}
