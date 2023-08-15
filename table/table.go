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
	numRows uint32
	pager   persistence.Pager
}

func GetNumRows(table *Table) uint32 {
	return table.numRows
}

// 行を挿入する
func (table *Table) InsertRow(row *Row) {
	// rs := table.getRowSlot(table.numRows)
	cursor := tableEnd(table)
	rs := cursorValue(&cursor)
	sourceBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(row))

	// ページに書き込む
	table.pager.InsertRow(rs, sourceBytes[:])
	table.numRows++
}

// 行を取得する
func (table *Table) GetRowByRowNum(rowNum uint32) Row {
	rs := persistence.GetRowSlot(rowNum)

	// ページから、Row構造体に書き込む
	row := &Row{}
	destinationBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(row))
	copy(destinationBytes[:], table.pager.GetRow(rs))

	return *row
}

type Cursor struct {
	table      *Table
	RowNum     uint32
	EndOfTable bool
}

func TableStart(table *Table) Cursor {
	var endOfTable bool
	if table.numRows == 0 {
		endOfTable = true
	} else {
		endOfTable = false
	}

	return Cursor{
		table:      table,
		RowNum:     0,
		EndOfTable: endOfTable,
	}
}

func tableEnd(table *Table) Cursor {
	return Cursor{
		table:      table,
		RowNum:     table.numRows,
		EndOfTable: true,
	}
}

func cursorValue(cursor *Cursor) persistence.RowSlot {
	return persistence.GetRowSlot(cursor.RowNum)
}

func CursorAdvance(cursor *Cursor) {
	cursor.RowNum += 1
	if cursor.RowNum >= cursor.table.numRows {
		cursor.EndOfTable = true
	}
}
