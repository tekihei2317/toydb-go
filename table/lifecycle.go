package table

import (
	"toydb-go/persistence"
)

func DbOpen(name string) (*Table, error) {
	pager, err := persistence.InitPager(name)

	if err != nil {
		return nil, err
	}

	// テーブルを初期化する
	table := Table{
		pager:       *pager,
		rootPageNum: 0,
	}

	return &table, nil
}

func DbClose(table *Table) {
	table.pager.FlushPages()
}
