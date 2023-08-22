package table

import (
	"fmt"
	"math"
	"strings"
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
	keyToInsert := uint32(row.Id)
	cursor := TableFind(table, keyToInsert)
	page := table.pager.GetPage(cursor.PageNum)

	persistence.LeafUtil.InsertCell(
		&table.pager,
		page,
		cursor.CellNum,
		uint32(row.Id),
		rowToBytes(row),
		table.rootPageNum,
	)
	return INSERT_SUCCESS
}

// 行を取得する
func (table *Table) GetRowByCursor(pageNum uint32, cellNum uint32) Row {
	// ページから、Row構造体に書き込む
	row := &Row{}
	destinationBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(row))

	page := table.pager.GetPage(pageNum)
	srcBytes := persistence.LeafUtil.GetCellValue(page, cellNum)
	copy(destinationBytes[:], srcBytes)

	return *row
}

type Cursor struct {
	table      *Table
	PageNum    uint32
	CellNum    uint32
	EndOfTable bool
}

func TableStart(table *Table) *Cursor {
	cursor := TableFind(table, 0)
	numCells := persistence.LeafUtil.GetNumCells(table.pager.GetPage(cursor.PageNum))
	if cursor.CellNum == numCells {
		cursor.EndOfTable = true
	}

	return cursor
}

// キー以上の最初のカーソルを返す
func TableFind(table *Table, key uint32) *Cursor {
	rootNode := table.pager.GetPage(table.rootPageNum)

	if persistence.NodeUtil.GetNodeType(rootNode) == persistence.NODE_LEAF {
		// リーフノードから探す
		return leafNodeFind(table, table.rootPageNum, key)
	} else {
		return internalNodeFind(table, table.rootPageNum, key)
	}
}

// 内部ノードから、リーフノードのキー以上の最初のカーソルを返す
func internalNodeFind(table *Table, pageNum uint32, key uint32) *Cursor {
	page := table.pager.GetPage(pageNum)
	numKeys := persistence.InternalUtil.GetNumKeys(page)

	// 内部ノードのキーを二分探索して、key以上の最初の要素が含まれる子ノードを見つける
	ng := -1
	ok := int(numKeys) + 1
	for ok-ng > 1 {
		index := (ng + ok) / 2
		var internalNodeKey uint32
		if uint32(index) == numKeys {
			// インデックスがnumKeysのキーは存在せず、ノードnumKeysは条件を満たすため、最大値を入れる
			internalNodeKey = math.MaxUint32
		} else {
			internalNodeKey = persistence.InternalUtil.GetKey(page, uint32(index))
		}

		if internalNodeKey < key {
			// ノードindexに含まれる値はkeyより小さいため、条件を満たさない
			ng = index
		} else {
			ok = index
		}
	}

	childPageNum := persistence.InternalUtil.GetChild(page, uint32(ok))
	childPage := table.pager.GetPage(childPageNum)

	switch persistence.NodeUtil.GetNodeType(childPage) {
	case persistence.NODE_INTERNAL:
		return internalNodeFind(table, childPageNum, key)
	case persistence.NODE_LEAF:
		return leafNodeFind(table, childPageNum, key)
	default:
		panic("Invalid node type")
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
		nextPageNum := persistence.LeafUtil.GetNextLeaf(page)

		if nextPageNum == 0 {
			cursor.EndOfTable = true
		} else {
			cursor.PageNum = nextPageNum
			cursor.CellNum = 0
		}
	}
}

func indent(level int) string {
	return strings.Repeat("  ", level)
}

// ノードを表示する
func PrintTree(table *Table, pageNum uint32, depth int) {
	node := table.pager.GetPage(pageNum)

	switch persistence.NodeUtil.GetNodeType(node) {
	case persistence.NODE_LEAF:
		numCells := persistence.LeafUtil.GetNumCells(node)
		fmt.Printf("%s- leaf (size %d)\n", indent(depth), numCells)
		for i := uint32(0); i < numCells; i++ {
			fmt.Printf("%s- %d\n", indent(depth+1), persistence.LeafUtil.GetCellKey(node, i))
		}
	case persistence.NODE_INTERNAL:
		numKeys := persistence.InternalUtil.GetNumKeys(node)
		fmt.Printf("%s- internal (size %d)\n", indent(depth), numKeys)

		if numKeys > 0 {
			for i := uint32(0); i < numKeys; i++ {
				// 子ノードを表示する
				childPageNum := persistence.InternalUtil.GetChild(node, i)
				PrintTree(table, childPageNum, depth+1)

				// キーを表示する
				fmt.Printf("%s- key %d\n", indent(depth+1), persistence.InternalUtil.GetKey(node, i))
			}
			// 一番右の子ノードを表示する
			childPageNum := persistence.InternalUtil.GetRightChild(node)
			PrintTree(table, childPageNum, depth+1)
		}
	}
}
