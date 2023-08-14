package main

import (
	"reflect"
	"testing"
)

func TestInsert(t *testing.T) {
	table := Table{numRows: 0}

	table.insertRow(&Row{
		id:       1,
		username: [32]byte{'t', 'e', 'k', 'i', 'h', 'e', 'i'},
		email:    [256]byte{'e', 'm', 'a', 'i', 'l', '@', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm'},
	})

	rs := table.getRowSlot(0)
	serializedRow := table.pager.getRow(rs)

	id := serializedRow[0:ID_SIZE]
	if !reflect.DeepEqual(id, []byte{1, 0, 0, 0, 0, 0, 0, 0}) {
		t.Errorf("invalid id. actual: %+v", id)
	}

	username := serializedRow[USERNAME_OFFSET:(USERNAME_OFFSET + USERNAME_SIZE)]
	expectedUsername := [32]byte{'t', 'e', 'k', 'i', 'h', 'e', 'i'}

	if !reflect.DeepEqual(username, expectedUsername[:]) {
		t.Errorf("invalid username. expected: %v, but got: %v", expectedUsername, username)
	}

	email := serializedRow[EMAIL_OFFSET : EMAIL_OFFSET+EMAIL_SIZE]
	expectedEmail := [256]byte{'e', 'm', 'a', 'i', 'l', '@', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm'}

	if !reflect.DeepEqual(email, expectedEmail[:]) {
		t.Errorf("invalid email. expected: %v, but got: %v", expectedEmail, email)
	}
}

func TestGetRow(t *testing.T) {
	table := Table{numRows: 0}

	row := Row{
		id:       1,
		username: [32]byte{'t', 'e', 'k', 'i', 'h', 'e', 'i'},
		email:    [256]byte{'e', 'm', 'a', 'i', 'l', '@', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm'},
	}

	table.insertRow(&row)

	fetchedRow := table.getRowByRowNum(0)
	if !reflect.DeepEqual(fetchedRow, row) {
		t.Errorf("invalid row. expected: %v, but got: %v", row, fetchedRow)
	}
}
