package persistence

import (
	"encoding/binary"
	"fmt"
	"math"
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
	oldMax := NodeUtil.getMaxKey(pager, oldNode)
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
	// TODO: 親ノードの番号を取得して設定する
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
		newMax := NodeUtil.getMaxKey(pager, oldNode)
		parent := pager.GetPage(parentPageNum)

		updateInternalNodeKey(parent, oldMax, newMax)
		internalNodeInsert(pager, parentPageNum, rightChildPageNum)
		return
	}
}

// ルートノードを分割したときに、新しいルートノードを作成する。
// 古いルートノード（左のノード）は、新しく作成したノードにコピーする。
// 新しく作成したルートノードには、左のノードのキーの最大値と、左右のノードのポインタ（ページ番号）をセットする。
func createNewRoot(pager *Pager, rootPageNum uint32, rightChildPageNum uint32) {
	root := pager.GetPage(rootPageNum)
	leftChild, leftChildPageNum := pager.GetNewPage()
	rightChild := pager.GetPage(rightChildPageNum)

	if NodeUtil.GetNodeType(root) == NODE_INTERNAL {
		initInternalNode(rightChild)
		initInternalNode(leftChild)
	}

	// ??
	if NodeUtil.GetNodeType(leftChild) == NODE_INTERNAL {
		var child *Page
		for i := 0; i < int(InternalUtil.GetNumKeys(leftChild)); i++ {
			childPage := InternalUtil.GetChild(leftChild, uint32(i))
			child = pager.GetPage(childPage)
			NodeUtil.setParent(child, leftChildPageNum)
		}
		// child = pager.GetPage(InternalUtil.GetRightChild(leftChild))
		// NodeUtil.setParent(child, leftChildPageNum)
	}

	// ルートノードを左のノードにコピーする
	copy(leftChild[:], root[:])
	NodeUtil.setNodeRoot(leftChild, false)

	// 新しいルートノードにデータをセットする
	initInternalNode(root)
	NodeUtil.setNodeRoot(root, true)
	InternalUtil.setNumKeys(root, 1)

	InternalUtil.setChild(root, 0, leftChildPageNum)
	leftChildMaxKey := NodeUtil.getMaxKey(pager, leftChild)
	InternalUtil.setKey(root, 0, leftChildMaxKey)
	InternalUtil.setRightChild(root, rightChildPageNum)

	// 左右のノードに親ノードを設定する
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

const INVALID_PAGE_NUM = math.MaxUint32

// 内部ノードを分割する。parentPageNumが分割する内部ノードで、childPageNumは新しく追加した右側の子ノード（？）
func internalNodeSplitAndInsert(pager *Pager, parentPageNum uint32, childPageNum uint32) {
	oldPageNum := parentPageNum
	oldNode := pager.GetPage(oldPageNum)
	oldMax := NodeUtil.getMaxKey(pager, oldNode)
	fmt.Println(oldPageNum, oldMax)

	child := pager.GetPage(childPageNum)
	childMax := NodeUtil.getMaxKey(pager, child)

	newNode, newPageNum := pager.GetNewPage()
	splittingRoot := isNodeRoot(oldNode)

	var parent *Page
	if splittingRoot {
		// oldNodeがルートノードだった場合
		createNewRoot(pager, oldPageNum, newPageNum)
		parent = pager.GetPage(oldPageNum)

		// oldPageNumが、新しいルートノードの左の子ノードを指すようにする
		oldPageNum := InternalUtil.GetChild(parent, 0)
		oldNode = pager.GetPage(oldPageNum)
	} else {
		// TODO:
		parent = pager.GetPage(NodeUtil.GetParent(oldNode))
		initInternalNode(newNode)
	}

	// newNodeに移動する
	curPageNum := InternalUtil.GetRightChild(oldNode)
	fmt.Println(curPageNum)
	cur := pager.GetPage(curPageNum)

	fmt.Println(newPageNum, curPageNum)
	internalNodeInsert(pager, newPageNum, curPageNum)
	NodeUtil.setParent(cur, newPageNum)
	InternalUtil.setRightChild(oldNode, INVALID_PAGE_NUM)

	for i := INTERNAL_NODE_MAX_CELLS - 1; i > INTERNAL_NODE_MAX_CELLS/2; i-- {
		curPageNum = InternalUtil.GetChild(oldNode, uint32(i))
		cur = pager.GetPage(curPageNum)
		internalNodeInsert(pager, newPageNum, curPageNum)

		NodeUtil.setParent(cur, newPageNum)
		InternalUtil.setNumKeys(oldNode, InternalUtil.GetNumKeys(oldNode)-1)
	}

	oldNumKeys := InternalUtil.GetNumKeys(oldNode)
	InternalUtil.setRightChild(oldNode, InternalUtil.GetChild(oldNode, oldNumKeys-1))
	InternalUtil.setNumKeys(oldNode, oldNumKeys-1)

	maxAfterSplit := NodeUtil.getMaxKey(pager, oldNode)
	var destinationPageNum uint32
	if childMax < maxAfterSplit {
		destinationPageNum = oldPageNum
	} else {
		destinationPageNum = newPageNum
	}

	internalNodeInsert(pager, destinationPageNum, childPageNum)
	NodeUtil.setParent(child, destinationPageNum)
	updateInternalNodeKey(parent, oldMax, NodeUtil.getMaxKey(pager, oldNode))

	if !splittingRoot {
		internalNodeInsert(pager, NodeUtil.GetParent(oldNode), newPageNum)
		NodeUtil.setParent(newNode, NodeUtil.GetParent(oldNode))
	}
}

// 新たに追加したノードとキーを、内部ノードの適切な位置に追加する
func internalNodeInsert(pager *Pager, parentPageNum uint32, childPageNum uint32) {
	parent := pager.GetPage(parentPageNum)
	child := pager.GetPage(childPageNum)
	childMaxKey := NodeUtil.getMaxKey(pager, child)
	index := internalNodeFindChild(parent, childMaxKey)

	originalNumKeys := InternalUtil.GetNumKeys(parent)
	if originalNumKeys >= INTERNAL_NODE_MAX_CELLS {
		// 内部ノードの容量がいっぱいの場合、分割する
		internalNodeSplitAndInsert(pager, parentPageNum, childPageNum)
		return
	}

	rightChildPageNum := InternalUtil.GetRightChild(parent)
	if rightChildPageNum == INVALID_PAGE_NUM {
		// ?
		InternalUtil.setRightChild(parent, childPageNum)
		return
	}

	rightChild := pager.GetPage(rightChildPageNum)
	InternalUtil.setNumKeys(parent, originalNumKeys+1)
	rightChildMaxKey := NodeUtil.getMaxKey(pager, rightChild)

	if childMaxKey > rightChildMaxKey {
		// 一番右側に挿入する場合
		// 一番右側のセルに、旧右childの値を設定する
		InternalUtil.setChild(parent, originalNumKeys, rightChildPageNum)
		InternalUtil.setKey(parent, originalNumKeys, rightChildMaxKey)

		// 新右childのページ番号を設定する
		InternalUtil.setRightChild(parent, childPageNum)
	} else {
		// indexに挿入できるように、ずらす
		for i := originalNumKeys; i > index; i-- {
			InternalUtil.setCell(parent, i, InternalUtil.GetCell(parent, i-1))
		}
		InternalUtil.setChild(parent, index, childPageNum)
		InternalUtil.setKey(parent, index, childMaxKey)
	}
}
