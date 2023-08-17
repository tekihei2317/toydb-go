package persistence

import (
	"encoding/binary"
	"fmt"
	"os"
)

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

func uint32ToBytes(v uint32) []byte {
	// binary.MaxVariantLen32 = 5の長さを確保しないと、クラッシュすることがあるらしい（4バイトなのになんでだろう）
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, v)
	return bytes
}

// リーフノードに関するユーティリティ関数
type leafUtil struct{}

var LeafUtil leafUtil

// ページのセルの数を返す
func (leafUtil) GetNumCells(page *Page) uint32 {
	numCellsBytes := page[LEAF_NODE_NUM_CELLS_OFFSET : LEAF_NODE_NUM_CELLS_OFFSET+LEAF_NODE_NUM_CELLS_SIZE]
	return binary.LittleEndian.Uint32(numCellsBytes)
}

// セルの個数をページに書き込む
func (leafUtil) WriteNumCells(page *Page, numCells uint32) {
	bytes := uint32ToBytes(numCells)

	copy(page[LEAF_NODE_NUM_CELLS_OFFSET:LEAF_NODE_NUM_CELLS_OFFSET+LEAF_NODE_NUM_CELLS_SIZE], bytes)
}

// セルの個数を1増やす
func (leafUtil) IncrementNumCells(page *Page) {
	numCells := LeafUtil.GetNumCells(page)
	LeafUtil.WriteNumCells(page, numCells+1)
}

// ノードにセルを挿入する
func (leafUtil) InsertCell(page *Page, cellNum uint32, key uint32, value []byte) {
	numCells := LeafUtil.GetNumCells(page)

	if numCells >= LEAF_NODE_MAX_CELLS {
		fmt.Printf("Need to implement splitting a leaf node.\n")
		os.Exit(1)
	}

	// セルの数を1増やす
	LeafUtil.IncrementNumCells(page)
	LeafUtil.WriteNodeKey(page, cellNum, key)
	LeafUtil.WriteNodeValue(page, cellNum, value)
}

// セルのキーをページに書き込む
func (leafUtil) WriteNodeKey(page *Page, cellNum uint32, key uint32) {
	start, end := LeafUtil.getKeyPos(cellNum)
	copy(page[start:end], uint32ToBytes(key))
}

// セルの値をページに書き込む
func (leafUtil) WriteNodeValue(page *Page, cellNum uint32, value []byte) {
	start, end := LeafUtil.getCellPos(cellNum)
	copy(page[start:end], value)
}

// リーフノードのセルの値を返す
func (leafUtil) GetCell(page *Page, cellNum uint32) []byte {
	start, end := LeafUtil.getCellPos(cellNum)
	return page[start:end]
}

// リーフノードのキーの値を返す
func (leafUtil) GetKey(page *Page, cellNum uint32) uint32 {
	start, end := LeafUtil.getKeyPos(cellNum)
	key := binary.LittleEndian.Uint32(page[start:end])
	return key
}

// リーフノードのセルの位置を返す
func (leafUtil) getCellPos(cellNum uint32) (uint32, uint32) {
	cellStart := LEAF_NODE_HEADER_SIZE + cellNum*LEAF_NODE_CELL_SIZE

	return cellStart, cellStart + LEAF_NODE_CELL_SIZE
}

// リーフノードのキーの位置を返す
func (leafUtil) getKeyPos(cellNum uint32) (uint32, uint32) {
	cellStart, _ := LeafUtil.getCellPos(cellNum)

	return cellStart, cellStart + LEAF_NODE_KEY_SIZE
}

// リーフノードの値の位置を返す
func (leafUtil) getValuePos(cellNum uint32) (uint32, uint32) {
	cellStart, _ := LeafUtil.getCellPos(cellNum)

	return cellStart + LEAF_NODE_KEY_SIZE, cellStart + LEAF_NODE_KEY_SIZE + LEAF_NODE_VALUE_SIZE
}
