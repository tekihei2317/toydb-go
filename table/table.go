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

func rowToBytes(row *Row) []byte {
	bytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(row))
	return bytes[:]
}

// 行を挿入する
func (table *Table) InsertRow(row *Row) {
	cursor := TableEnd(table)
	page := table.pager.GetPage(cursor.PageNum)

	persistence.LeafUtil.InsertCell(page, cursor.CellNum, uint32(row.Id), rowToBytes(row))
}

// 行を取得する
func (table *Table) GetRowByCursor(pageNum uint32, cellNum uint32) Row {
	// ページから、Row構造体に書き込む
	row := &Row{}
	destinationBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(row))

	page := table.pager.GetPage(pageNum)
	srcBytes := persistence.LeafUtil.GetCell(page, cellNum)
	copy(destinationBytes[:], srcBytes)

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
	numCells := persistence.LeafUtil.GetNumCells(table.pager.GetPage(table.rootPageNum))
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

func TableEnd(table *Table) Cursor {
	numCells := persistence.LeafUtil.GetNumCells(table.pager.GetPage(table.rootPageNum))

	return Cursor{
		table:      table,
		PageNum:    table.rootPageNum,
		CellNum:    numCells,
		EndOfTable: true,
	}
}

// カーソルを1つ進める
func CursorAdvance(cursor *Cursor) {
	page := cursor.table.pager.GetPage(cursor.PageNum)
	numCells := persistence.LeafUtil.GetNumCells(page)
	cursor.CellNum += 1

	if cursor.CellNum >= numCells {
		cursor.EndOfTable = true
	}
}
