package main

import (
	"os"
)

func dbOpen(name string) (*Table, error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	// ページャーを初期化する
	pager := Pager{
		file:  f,
		pages: [TABLE_MAX_PAGES]*Page{},
	}

	// 行数を計算する。ファイルサイズと1行のサイズから計算できる。
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	numRows := uint32(fi.Size() / int64(ROW_SIZE))

	// テーブルを初期化する
	table := Table{
		numRows: numRows,
		pager:   pager,
	}

	return &table, nil
}

func dbClose(table *Table) {
	table.pager.flushPages(table.numRows)
}
