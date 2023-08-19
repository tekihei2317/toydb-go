# memo

## Part 9

リーフノードに挿入する位置を二分探索で求める。また、キーが重複した場合にエラーになるようにする。

二分探索では、キー以上となる最初のキーの位置を見つける。存在しなかったら、最後のキーの一つ後ろの位置を返す。

```go
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
```

挿入する位置を見つけた後に後ろにずらす処理みたいなのが必要なのかなと思った。今の実装では、1→3→2と挿入した時に、3を2で上書きして`(1, 2, empty)`みたいな感じになった。

ソースコードを読んでみたら、ずらす部分の処理を実装するのを飛ばしていたのが原因だった。なので`cellNum < numCells`の部分の処理を追加した。

```go
// ノードにセルを挿入する
func (leafUtil) InsertCell(page *Page, cellNum uint32, key uint32, value []byte) {
	numCells := LeafUtil.GetNumCells(page)

	if numCells >= LEAF_NODE_MAX_CELLS {
		fmt.Printf("Need to implement splitting a leaf node.\n")
		os.Exit(1)
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
```
最大でページサイズ（4096バイト）くらいのコピーの処理になるので、ページのフォーマットを変えると改善できそうだなと思った。例えば、行の挿入は末尾にして、キーの昇順にセル番号を持っておくとか。

具体的には、セルに(1, 4, 2)という順番でキーが追加されたなら、(0, 2, 1)みたいに持つ。3を挿入する場合は、セルは(1, 4, 2, 3)、セル番号は(0, 2, 3, 1)になる。これならコピーは最大でセルの数程度になる。

## Part 10

まだツリーになっていないので、リーフノードを分割できるようにする。一旦、ツリーにするために必要なことを書き出してみる。

- 再帰的にノードを探索する
- リーフノードがいっぱいになった時に、リーフノードを分割して親ノードを更新する
	- 親ノードの更新を再帰的に行う

Part 10では、ノードを2つに分割するところと、新しいルートノードを作る部分が難しい。

新しいルートノードを作るとき、新しいページをルートノードにするのではないのが気になった。新しいページに左のノードを移動して、左のノードをルートノードにしていた。
