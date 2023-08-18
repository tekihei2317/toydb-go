package table

import (
	"fmt"
	"os"
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

type InsertResult int

const (
	INSERT_SUCCESS InsertResult = iota + 1
	INSERT_TABLE_FULL
	INSERT_DUPLICATE_KEY
)

// 行を挿入する
func (table *Table) InsertRow(row *Row) InsertResult {
	// 現時点では、rootPageがリーフノードであるとする
	page := table.pager.GetPage(table.rootPageNum)
	numCells := persistence.LeafUtil.GetNumCells(page)
	if numCells >= persistence.LEAF_NODE_MAX_CELLS {
		return INSERT_TABLE_FULL
	}

	keyToInsert := uint32(row.Id)
	cursor := TableFind(table, keyToInsert)

	if cursor.CellNum < numCells {
		keyAtCursor := persistence.LeafUtil.GetCellKey(page, cursor.CellNum)
		if keyAtCursor == keyToInsert {
			return INSERT_DUPLICATE_KEY
		}
	}

	persistence.LeafUtil.InsertCell(page, cursor.CellNum, uint32(row.Id), rowToBytes(row))
	return INSERT_SUCCESS
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

// キー以上の最初のカーソルを返す
func TableFind(table *Table, key uint32) *Cursor {
	rootNode := table.pager.GetPage(table.rootPageNum)

	if persistence.LeafUtil.GetNodeType(rootNode) == persistence.NODE_LEAF {
		// リーフノードから探す
		return leafNodeFind(table, table.rootPageNum, key)
	} else {
		fmt.Println("Need to implement searching an internal node")
		os.Exit(1)
		return &Cursor{}
	}
}

// リーフノードから、キー以上の最初のカーソルを返す
func leafNodeFind(table *Table, pageNum uint32, key uint32) *Cursor {
	node := table.pager.GetPage(pageNum)
	numCells := persistence.LeafUtil.GetNumCells(node)
	cursor := &Cursor{
		table:   table,
		PageNum: pageNum,
		CellNum: 0,
	}

	// キー以上の最初のインデックスを探す
	ng, ok := -1, int(numCells)
	for ok-ng > 1 {
		i := (ok + ng) / 2
		keyAtI := persistence.LeafUtil.GetCellKey(node, uint32(i))

		if keyAtI >= key {
			ok = i
		} else {
			ng = i
		}
	}
	cursor.CellNum = uint32(ok)

	return cursor
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

// リーフノードを表示する
func PrintLeafNode(table *Table) {
	page := table.pager.GetPage(0)
	numCells := persistence.LeafUtil.GetNumCells(page)
	fmt.Printf("leaf (size %d)\n", numCells)

	for i := uint32(0); i < numCells; i++ {
		key := persistence.LeafUtil.GetCellKey(page, i)
		fmt.Printf("  - %d : %d\n", i, key)
	}
}
