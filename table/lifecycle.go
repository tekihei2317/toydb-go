package table

import (
	"toydb-go/persistence"
)

func DbOpen(name string) (*Table, error) {
	pager, numRows, err := persistence.InitPager(name)

	if err != nil {
		return nil, err
	}

	// テーブルを初期化する
	table := Table{
		numRows: numRows,
		pager:   *pager,
	}

	return &table, nil
}

func DbClose(table *Table) {
	table.pager.FlushPages(table.numRows)
}
