package table

import (
	"reflect"
	"testing"
)

func TestInsertAndGetRow(t *testing.T) {
	table := Table{}

	row := Row{
		Id:       1,
		Username: [32]byte{'t', 'e', 'k', 'i', 'h', 'e', 'i'},
		Email:    [256]byte{'e', 'm', 'a', 'i', 'l', '@', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm'},
	}

	table.InsertRow(&row)

	fetchedRow := table.GetRowByCursor(0, 0)
	if !reflect.DeepEqual(fetchedRow, row) {
		t.Errorf("invalid row. expected: %v, but got: %v", row, fetchedRow)
	}
}
