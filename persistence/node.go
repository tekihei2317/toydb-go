package persistence

import (
	"encoding/binary"
	"fmt"
	"os"
)

// Common Node Header Layout
const (
	// ノードの種類（internal / leaf）
	NODE_TYPE_SIZE   = 1
	NODE_TYPE_OFFSET = 0
	// ルートノードかどうか
	IS_ROOT_SIZE   = 1
	IS_ROOT_OFFSET = NODE_TYPE_SIZE
	// 親ノードへのポインタ
	PARENT_POINTER_SIZE     = 4
	PARENT_POINTER_OFFSET   = IS_ROOT_OFFSET + IS_ROOT_SIZE
	COMMON_NODE_HEADER_SIZE = NODE_TYPE_SIZE + IS_ROOT_SIZE + PARENT_POINTER_SIZE
)

// Leaf Node Header Layout(Common Header + Cells size)
const (
	// ノードに含まれるセル（key/valueのペア）の数
	LEAF_NODE_NUM_CELLS_SIZE   = 4
	LEAF_NODE_NUM_CELLS_OFFSET = COMMON_NODE_HEADER_SIZE
	// 右隣のリーフノードのページ番号
	LEAF_NODE_NEXT_LEAF_SIZE   = 4
	LEAF_NODE_NEXT_LEAF_OFFSET = LEAF_NODE_NUM_CELLS_OFFSET + LEAF_NODE_NUM_CELLS_SIZE
	LEAF_NODE_HEADER_SIZE      = COMMON_NODE_HEADER_SIZE + LEAF_NODE_NUM_CELLS_SIZE + LEAF_NODE_NEXT_LEAF_SIZE
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

	// MAX + 1個になったとき、右は半分（切捨て）、左は残りに分割する
	LEAF_NODE_RIGHT_SPLIT_COUNT = (LEAF_NODE_MAX_CELLS + 1) / 2
	LEAF_NODE_LEFT_SPLIT_COUNT  = (LEAF_NODE_MAX_CELLS + 1) - LEAF_NODE_RIGHT_SPLIT_COUNT
)

// Internal Node Header Layout
const (
	// 内部ノードに含まれるキーの数
	INTERNAL_NODE_NUM_KEYS_SIZE   = 4
	INTERNAL_NODE_NUM_KEYS_OFFSET = COMMON_NODE_HEADER_SIZE
	// 右のノード（兄弟？）へのポインタ
	INTERNAL_NODE_RIGHT_CHILD_SIZE   = 4
	INTERNAL_NODE_RIGHT_CHILD_OFFSET = INTERNAL_NODE_NUM_KEYS_OFFSET + INTERNAL_NODE_NUM_KEYS_SIZE
	INTERNAL_NODE_HEADER_SIZE        = COMMON_NODE_HEADER_SIZE + INTERNAL_NODE_NUM_KEYS_SIZE + INTERNAL_NODE_RIGHT_CHILD_SIZE
)

// Internal Node Body Layout
const (
	INTERNAL_NODE_KEY_SIZE   = 4
	INTERNAL_NODE_CHILD_SIZE = 4
	INTERNAL_NODE_CELL_SIZE  = INTERNAL_NODE_CHILD_SIZE + INTERNAL_NODE_KEY_SIZE
)

func initLeafNode(node *Page) {
	NodeUtil.setNodeType(node, NODE_LEAF)
	NodeUtil.setNodeRoot(node, false)
	LeafUtil.setNextLeaf(node, 0) // 0は隣のノードがないことを表す
}

func initInternalNode(node *Page) {
	NodeUtil.setNodeType(node, NODE_INTERNAL)
	NodeUtil.setNodeRoot(node, false)
	InternalUtil.setNumKeys(node, 0)
}

// ノードに関するユーティリティ関数
type nodeUtil struct{}

var NodeUtil nodeUtil

func (nodeUtil) setNodeType(page *Page, nodeType NodeType) {
	bytes := []byte{byte(nodeType)}
	copy(page[NODE_TYPE_OFFSET:NODE_TYPE_OFFSET+NODE_TYPE_SIZE], bytes)
}

func boolToByte(v bool) byte {
	if v {
		return 1
	}
	return 0
}

func byteToBool(v byte) bool {
	if v == 1 {
		return true
	}
	return false
}

// ノードの種類（INTERNAL、LEAF）を返す
func (nodeUtil) GetNodeType(page *Page) NodeType {
	bytes := page[NODE_TYPE_OFFSET : NODE_TYPE_OFFSET+NODE_TYPE_SIZE]
	return NodeType(bytes[0])
}

// ルートノードかどうかを設定する
func (nodeUtil) setNodeRoot(page *Page, isRoot bool) {
	bytes := []byte{boolToByte(isRoot)}
	copy(page[IS_ROOT_OFFSET:IS_ROOT_OFFSET+IS_ROOT_SIZE], bytes)
}

// ルートノードかどうかを返す
func isNodeRoot(page *Page) bool {
	v := page[IS_ROOT_OFFSET : IS_ROOT_OFFSET+IS_ROOT_SIZE][0]
	return byteToBool(v)
}

// 親ノードを設定する
func (nodeUtil) setParent(page *Page, parentPageNum uint32) {
	bytes := uint32ToBytes(parentPageNum)
	copy(page[PARENT_POINTER_OFFSET:PARENT_POINTER_OFFSET+PARENT_POINTER_SIZE], bytes)
}

// 親ノードを取得する
func (nodeUtil) GetParent(page *Page) uint32 {
	bytes := page[PARENT_POINTER_OFFSET : PARENT_POINTER_OFFSET+PARENT_POINTER_SIZE]
	return binary.LittleEndian.Uint32(bytes)
}

// ノードに含まれる最大のキーを返す
func (nodeUtil) getMaxKey(page *Page) uint32 {
	nodeType := NodeUtil.GetNodeType(page)

	var key uint32
	if nodeType == NODE_INTERNAL {
		// 最後のキーを返す
		key = InternalUtil.GetKey(page, InternalUtil.GetNumKeys(page)-1)
	} else {
		// 最後のセルのキーを返す
		key = LeafUtil.GetCellKey(page, LeafUtil.GetNumCells(page)-1)
	}
	return key
}

// 内部ノードに関するユーティリティ関数
type internalUtil struct{}

var InternalUtil internalUtil

// 内部ノードに含まれるキーの数を返す
func (internalUtil) GetNumKeys(page *Page) uint32 {
	bytes := page[INTERNAL_NODE_NUM_KEYS_OFFSET : INTERNAL_NODE_NUM_KEYS_OFFSET+INTERNAL_NODE_NUM_KEYS_SIZE]
	return binary.LittleEndian.Uint32(bytes)
}

// 内部ノードに含まれるキーの数を設定する
func (internalUtil) setNumKeys(page *Page, numKeys uint32) {
	bytes := uint32ToBytes(numKeys)
	copy(page[INTERNAL_NODE_NUM_KEYS_OFFSET:INTERNAL_NODE_NUM_KEYS_OFFSET+INTERNAL_NODE_NUM_KEYS_SIZE], bytes)
}

// 一番右の子供のページ番号？を返す
func (internalUtil) GetRightChild(page *Page) uint32 {
	bytes := page[INTERNAL_NODE_RIGHT_CHILD_OFFSET : INTERNAL_NODE_RIGHT_CHILD_OFFSET+INTERNAL_NODE_RIGHT_CHILD_SIZE]
	return binary.LittleEndian.Uint32(bytes)
}

// 一番右の子供のページ番号？を設定する
func (internalUtil) setRightChild(page *Page, rightChildPageNum uint32) {
	bytes := uint32ToBytes(rightChildPageNum)
	copy(page[INTERNAL_NODE_RIGHT_CHILD_OFFSET:INTERNAL_NODE_RIGHT_CHILD_OFFSET+INTERNAL_NODE_RIGHT_CHILD_SIZE], bytes)
}

// 子供のページ番号を取得する
func (internalUtil) GetChild(page *Page, cellNum uint32) uint32 {
	numKeys := InternalUtil.GetNumKeys(page)

	if cellNum > numKeys {
		fmt.Printf("Tried to get cellNum %d > numKeys %d\n", cellNum, numKeys)
		os.Exit(1)
		return 0
	} else if cellNum == numKeys {
		return InternalUtil.GetRightChild(page)
	} else {
		start, end := InternalUtil.getChildPos(cellNum)
		return binary.LittleEndian.Uint32(page[start:end])
	}
}

// 子供のページ番号を設定する
func (internalUtil) setChild(page *Page, cellNum uint32, pageNum uint32) {
	numKeys := InternalUtil.GetNumKeys(page)

	// インデックスがnumKeysまでのchildがある
	if cellNum > numKeys {
		fmt.Printf("Tried to set cellNum %d > numKeys %d\n", cellNum, numKeys)
		os.Exit(1)
	} else if cellNum == numKeys {
		// ここの処理必要なのかな
		fmt.Printf("Tried to set cellNum %d = numKeys %d\n", cellNum, numKeys)
		os.Exit(1)
	} else {
		start, end := InternalUtil.getChildPos(cellNum)
		copy(page[start:end], uint32ToBytes(pageNum))
	}
}

// キーを取得する
func (internalUtil) GetKey(page *Page, cellNum uint32) uint32 {
	start, end := InternalUtil.getKeyPos(cellNum)
	return binary.LittleEndian.Uint32(page[start:end])
}

// キーを設定する
func (internalUtil) setKey(page *Page, cellNum uint32, key uint32) {
	start, end := InternalUtil.getKeyPos(cellNum)
	copy(page[start:end], uint32ToBytes(key))
}

// セルの位置を返す
func (internalUtil) getCellPos(cellNum uint32) (uint32, uint32) {
	start := INTERNAL_NODE_HEADER_SIZE + INTERNAL_NODE_CELL_SIZE*cellNum
	end := start + INTERNAL_NODE_CELL_SIZE
	return start, end
}

// キーの位置を返す
func (internalUtil) getKeyPos(cellNum uint32) (uint32, uint32) {
	start, _ := InternalUtil.getCellPos(cellNum)
	return start, start + INTERNAL_NODE_KEY_SIZE
}

// 子供の位置を返す
func (internalUtil) getChildPos(cellNum uint32) (uint32, uint32) {
	start, _ := InternalUtil.getCellPos(cellNum)
	childStart := start + INTERNAL_NODE_KEY_SIZE
	return childStart, childStart + INTERNAL_NODE_CHILD_SIZE
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

// 右隣のリーフノードのページ番号を返す
func (leafUtil) GetNextLeaf(page *Page) uint32 {
	bytes := page[LEAF_NODE_NEXT_LEAF_OFFSET : LEAF_NODE_NEXT_LEAF_OFFSET+LEAF_NODE_NEXT_LEAF_SIZE]
	return binary.LittleEndian.Uint32(bytes)
}

// 右隣のリーフノードのページ番号を設定する
func (leafUtil) setNextLeaf(page *Page, pageNum uint32) {
	bytes := uint32ToBytes(pageNum)
	copy(page[LEAF_NODE_NEXT_LEAF_OFFSET:LEAF_NODE_NEXT_LEAF_OFFSET+LEAF_NODE_NEXT_LEAF_SIZE], bytes)
}

// セルの個数を1増やす
func (leafUtil) IncrementNumCells(page *Page) {
	numCells := LeafUtil.GetNumCells(page)
	LeafUtil.WriteNumCells(page, numCells+1)
}

// ノードにセルを挿入する
func (leafUtil) InsertCell(pager *Pager, page *Page, cellNum uint32, key uint32, value []byte, rootPageNum uint32) {
	numCells := LeafUtil.GetNumCells(page)

	if numCells >= LEAF_NODE_MAX_CELLS {
		leafNodeSplitAndInsert(pager, page, cellNum, key, value, rootPageNum)
		return
	}

	if cellNum < numCells {
		// cellNumに挿入できるように、それより後ろにあるセルを1つずつずらす
		for i := numCells; i > cellNum; i-- {
			toStart, toEnd := LeafUtil.getCellPos(i)
			copy(page[toStart:toEnd], LeafUtil.GetCell(page, i-1))
		}
	}
	LeafUtil.IncrementNumCells(page)
	LeafUtil.WriteCellKey(page, cellNum, key)
	LeafUtil.WriteCellValue(page, cellNum, value)
}

// リーフノードのセル（key + value）を返す
func (leafUtil) GetCell(page *Page, cellNum uint32) []byte {
	start, end := LeafUtil.getCellPos(cellNum)
	return page[start:end]
}

// リーフノードにセルを書き込む
func (leafUtil) WriteCell(page *Page, cellNum uint32, cell []byte) {
	start, end := LeafUtil.getCellPos(cellNum)
	copy(page[start:end], cell)
}

// セルのキーの値を返す
func (leafUtil) GetCellKey(page *Page, cellNum uint32) uint32 {
	start, end := LeafUtil.getKeyPos(cellNum)
	key := binary.LittleEndian.Uint32(page[start:end])
	return key
}

// セルのキーをページに書き込む
func (leafUtil) WriteCellKey(page *Page, cellNum uint32, key uint32) {
	start, end := LeafUtil.getKeyPos(cellNum)
	copy(page[start:end], uint32ToBytes(key))
}

// セルのvalueを返す
func (leafUtil) GetCellValue(page *Page, cellNum uint32) []byte {
	start, end := LeafUtil.getValuePos(cellNum)
	return page[start:end]
}

// セルのvalueをページに書き込む
func (leafUtil) WriteCellValue(page *Page, cellNum uint32, value []byte) {
	start, end := LeafUtil.getValuePos(cellNum)
	copy(page[start:end], value)
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
