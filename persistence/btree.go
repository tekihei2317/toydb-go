package persistence

import (
	"encoding/binary"
)

type NodeType uint8

const (
	NODE_INTERNAL NodeType = iota
	NODE_LEAF
)

func uint32ToBytes(v uint32) []byte {
	// binary.MaxVariantLen32 = 5の長さを確保しないと、クラッシュすることがあるらしい（4バイトなのになんでだろう）
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, v)
	return bytes
}

// リーフノードを分割してから、新しいセルを挿入する。
// リーフノードは、左と右に半分ずつ分割する。pageは分割するページ。
func leafNodeSplitAndInsert(pager *Pager, page *Page, cellNum uint32, key uint32, value []byte, rootPageNum uint32) {
	oldNode := page
	oldMax := NodeUtil.getMaxKey(oldNode)
	// 新しいページを取得する（右）
	newNode, rightChildPageNum := pager.GetNewPage()
	initLeafNode(newNode)

	// LEAF_NODE_MAX_CELLS+1個のセルを、左右に分割する
	for index := LEAF_NODE_MAX_CELLS; index >= 0; index-- {
		i := uint32(index)
		var dstNode *Page
		if i >= LEAF_NODE_LEFT_SPLIT_COUNT {
			dstNode = newNode
		} else {
			dstNode = oldNode
		}
		dstCellNum := i % LEAF_NODE_LEFT_SPLIT_COUNT

		if i == cellNum {
			// 挿入する位置の場合は、挿入する
			LeafUtil.WriteCellKey(dstNode, dstCellNum, key)
			LeafUtil.WriteCellValue(dstNode, dstCellNum, value)
		} else if i > cellNum {
			// 挿入する位置より後ろの場合は、コピー元は元のページのi-1番目
			LeafUtil.WriteCell(dstNode, dstCellNum, LeafUtil.GetCell(page, i-1))
		} else {
			// 挿入する位置より前の場合は、コピー元は元のページのi番目
			LeafUtil.WriteCell(dstNode, dstCellNum, LeafUtil.GetCell(page, i))
		}
	}
	LeafUtil.WriteNumCells(oldNode, LEAF_NODE_LEFT_SPLIT_COUNT)
	LeafUtil.WriteNumCells(newNode, LEAF_NODE_RIGHT_SPLIT_COUNT)
	// TODO: 親ノードの番号を取得して設定する（取れるかな）
	// NodeUtil.setParent(newNode, )
	LeafUtil.setNextLeaf(newNode, LeafUtil.GetNextLeaf(oldNode))
	LeafUtil.setNextLeaf(oldNode, rightChildPageNum)

	if isNodeRoot(oldNode) {
		// 新しいルートノードを作成する
		createNewRoot(pager, rootPageNum, rightChildPageNum)
		return
	} else {
		// 分割したリーフノードがルートではない場合は、親ノードを更新する
		parentPageNum := NodeUtil.GetParent(oldNode)
		newMax := NodeUtil.getMaxKey(oldNode)
		parent := pager.GetPage(parentPageNum)

		updateInternalNodeKey(parent, oldMax, newMax)
		return
	}
}

// ルートノードを分割したときに、新しいルートノードを作成する。
// 古いルートノード（左のノード）は、新しく作成したノードにコピーする。
// 新しく作成したルートノードには、左のノードのキーの最大値と、左右のノードのポインタ（ページ番号）をセットする。
func createNewRoot(pager *Pager, rootPageNum uint32, rightChildPageNum uint32) {
	root := pager.GetPage(rootPageNum)
	leftChild, leftChildPageNum := pager.GetNewPage()

	// ルートノードを左のノードにコピーする
	copy(leftChild[:], root[:])
	NodeUtil.setNodeRoot(leftChild, false)

	// 新しいルートノードにデータをセットする
	initInternalNode(root)
	NodeUtil.setNodeRoot(root, true)
	InternalUtil.setNumKeys(root, 1)

	InternalUtil.setChild(root, 0, leftChildPageNum)
	leftChildMaxKey := NodeUtil.getMaxKey(leftChild)
	InternalUtil.setKey(root, 0, leftChildMaxKey)
	InternalUtil.setRightChild(root, rightChildPageNum)

	rightChild := pager.GetPage(rightChildPageNum)
	NodeUtil.setParent(leftChild, rootPageNum)
	NodeUtil.setParent(rightChild, rootPageNum)
}

// リーフノードを分割した後に、内部ノードのキーを更新する
func updateInternalNodeKey(node *Page, oldKey uint32, newKey uint32) {
	oldChildIndex := internalNodeFindChild(node, oldKey)
	// oldChildIndexが、nodeのキーの数と同じだった場合は大丈夫？
	InternalUtil.setKey(node, oldChildIndex, newKey)
}

// キーが含まれる子ノードのインデックスを返す
func internalNodeFindChild(node *Page, key uint32) uint32 {
	numKeys := InternalUtil.GetNumKeys(node)

	// 内部ノードのキーを二分探索して、key以上の最初の要素が含まれる子ノードを見つける
	ng := -1
	ok := int(numKeys)
	for ok-ng > 1 {
		index := (ng + ok) / 2
		internalNodeKey := InternalUtil.GetKey(node, uint32(index))

		if internalNodeKey < key {
			// ノードindexに含まれる値internalNodeKey以下である。つまりkeyより小さいため、条件を満たさない。
			ng = index
		} else {
			ok = index
		}
	}
	return uint32(ok)
}
