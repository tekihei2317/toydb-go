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

// func leaf_node_num_cells(void* node) {
//   return node + LEAF_NODE_NUM_CELLS_OFFSET;
// }

// ページのセルの数を返す
func getLeafNodeNumCells(page *Page) uint32 {
	numCellsBytes := page[LEAF_NODE_NUM_CELLS_OFFSET : LEAF_NODE_NUM_CELLS_OFFSET+LEAF_NODE_NUM_CELLS_SIZE]
	return binary.LittleEndian.Uint32(numCellsBytes)
}

// リーフノードのセルの位置を返す
func leafNodeCell(cellNum int) (int, int) {
	cellStart := LEAF_NODE_HEADER_SIZE + cellNum*LEAF_NODE_CELL_SIZE

	return cellStart, cellStart + LEAF_NODE_CELL_SIZE
}

// リーフノードのキーの位置を返す
func leafNodeKey(cellNum int) (int, int) {
	cellStart, _ := leafNodeCell(cellNum)

	return cellStart, cellStart + LEAF_NODE_KEY_SIZE
}

// リーフノードの値の位置を返す
func leafNodeValue(cellNum int) (int, int) {
	cellStart, _ := leafNodeCell(cellNum)

	return cellStart + LEAF_NODE_KEY_SIZE, cellStart + LEAF_NODE_KEY_SIZE + LEAF_NODE_VALUE_SIZE
}
