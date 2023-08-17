package persistence

import "encoding/binary"

type NodeType int

const (
	NODE_INTERNAL NodeType = iota
	NODE_LEAF
)

// Common Node Header Layout
const (
	NODE_TYPE_SIZE          = 1
	NODE_TYPE_OFFSET        = 0
	IS_ROOT_SIZE            = 1
	IS_ROOT_OFFSET          = NODE_TYPE_SIZE
	PARENT_POINTER_SIZE     = 4
	PARENT_POINTER_OFFSET   = IS_ROOT_OFFSET + IS_ROOT_SIZE
	COMMON_NODE_HEADER_SIZE = NODE_TYPE_SIZE + IS_ROOT_SIZE + PARENT_POINTER_SIZE
)

// Leaf Node Header Layout(Common Header + Cells size)
const (
	LEAF_NODE_NUM_CELLS_SIZE   = 4
	LEAF_NODE_NUM_CELLS_OFFSET = COMMON_NODE_HEADER_SIZE
	LEAF_NODE_HEADER_SIZE      = COMMON_NODE_HEADER_SIZE + LEAF_NODE_NUM_CELLS_SIZE
)

// Leaf Node Body Layout
const (
	LEAF_NODE_KEY_OFFSET   = 0
	LEAF_NODE_KEY_SIZE     = 4
	LEAF_NODE_VALUE_OFFSET = LEAF_NODE_KEY_SIZE
	LEAF_NODE_VALUE_SIZE   = ROW_SIZE
	LEAF_NODE_CELL_SIZE    = LEAF_NODE_KEY_SIZE + LEAF_NODE_VALUE_SIZE

	LEAF_NODE_SPACE_FOR_CELLS = PAGE_SIZE - LEAF_NODE_HEADER_SIZE
	LEAF_NODE_MAX_CELLS       = LEAF_NODE_SPACE_FOR_CELLS / LEAF_NODE_CELL_SIZE
)

// ページのセルの数を返す
func getLeafNodeNumCells(page *Page) uint32 {
	numCellsBytes := page[LEAF_NODE_NUM_CELLS_OFFSET : LEAF_NODE_NUM_CELLS_OFFSET+LEAF_NODE_NUM_CELLS_SIZE]
	return binary.LittleEndian.Uint32(numCellsBytes)
}

func uint32ToBytes(v uint32) []byte {
	// binary.MaxVariantLen32 = 5の長さを確保しないと、クラッシュすることがあるらしい（4バイトなのになんでだろう）
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, v)
	return bytes
}

// セルの数をバイトにエンコードして、ページに書き込む
func writeLeafNodeNumCells(page *Page, numCells uint32) {
	bytes := uint32ToBytes(numCells)

	copy(page[LEAF_NODE_NUM_CELLS_OFFSET:LEAF_NODE_NUM_CELLS_OFFSET+LEAF_NODE_NUM_CELLS_SIZE], bytes)
}

// セルのキーを書き込む
func writeLeafNodeKey(page *Page, cellNum uint32, key uint32) {
	start, end := leafNodeKey(cellNum)
	copy(page[start:end], uint32ToBytes(key))
}

func writeLeafNodeValue(page *Page, cellNum uint32, value []byte) {
	start, end := leafNodeCell(cellNum)
	copy(page[start:end], value)
}

// リーフノードのセルの位置を返す
func leafNodeCell(cellNum uint32) (uint32, uint32) {
	cellStart := LEAF_NODE_HEADER_SIZE + cellNum*LEAF_NODE_CELL_SIZE

	return cellStart, cellStart + LEAF_NODE_CELL_SIZE
}

// リーフノードのキーの位置を返す
func leafNodeKey(cellNum uint32) (uint32, uint32) {
	cellStart, _ := leafNodeCell(cellNum)

	return cellStart, cellStart + LEAF_NODE_KEY_SIZE
}

// リーフノードの値の位置を返す
func leafNodeValue(cellNum uint32) (uint32, uint32) {
	cellStart, _ := leafNodeCell(cellNum)

	return cellStart + LEAF_NODE_KEY_SIZE, cellStart + LEAF_NODE_KEY_SIZE + LEAF_NODE_VALUE_SIZE
}
