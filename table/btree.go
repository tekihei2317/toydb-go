package table

import (
	"fmt"
	"os"
	"toydb-go/persistence"
	"unsafe"
)

func rowToBytes(row *Row) []byte {
	bytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(row))
	return bytes[:]
}

// リーフノードに挿入する
func leafNodeInsert(cursor *Cursor, key uint32, value *Row) {
	numCells := persistence.GetNumCells(&cursor.table.pager, cursor.PageNum)

	if numCells >= persistence.LEAF_NODE_MAX_CELLS {
		fmt.Printf("Need to implement splitting a leaf node.\n")
		os.Exit(1)
	}

	// セルの数を1増やす
	persistence.IncrementNumCells(&cursor.table.pager, cursor.PageNum)
	// セルのキーを書き込む
	persistence.WriteLeafNodeKey(&cursor.table.pager, cursor.PageNum, key)
	// セルの値を書き込む
	persistence.WriteLeafNodeValue(&cursor.table.pager, cursor.PageNum, rowToBytes(value))
}
