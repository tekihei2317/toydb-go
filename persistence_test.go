package main

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestPointer(t *testing.T) {
	// unsafe.SizeOfは、渡した値のサイズをバイトで返す

	if unsafe.Sizeof(int(0)) != 8 {
		t.Error("Error 1")
	}

	if unsafe.Sizeof(int32(0)) != 4 {
		t.Error("Error 2")
	}

	// バイトの32要素の配列
	if unsafe.Sizeof([32]byte{}) != 32 {
		t.Error("Error 3")
	}

	// 文字列1文字は、16ビット=4バイトのサイズ（絵文字に合わせているのかな）
	if unsafe.Sizeof("a") != 16 {
		t.Errorf("Error 4, actual %d\n", unsafe.Sizeof("a"))
	}

	if unsafe.Sizeof("あ") != 16 {
		t.Errorf("Error 5, actual %d\n", unsafe.Sizeof("あ"))
	}
}

func TestSerializeRow(t *testing.T) {
	if unsafe.Sizeof(Row{}) != 8+32+256 {
		t.Errorf("Error 1, actual: %d\n", unsafe.Sizeof(Row{}))
	}

	source := Row{
		id:       1,
		username: [32]byte{'t', 'e', 'k', 'i', 'h', 'e', 'i'},
		email:    [256]byte{'e', 'm', 'a', 'i', 'l', '@', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm'},
	}

	var serializedRow [ROW_SIZE]byte
	serializeRow(&source, serializedRow[:])

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

func TestDeserializeRow(t *testing.T) {
	source := Row{
		id:       1,
		username: [32]byte{'t', 'e', 'k', 'i', 'h', 'e', 'i'},
		email:    [256]byte{'e', 'm', 'a', 'i', 'l', '@', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm'},
	}

	var serializedRow = make([]byte, ROW_SIZE)
	serializeRow(&source, serializedRow[:])

	deserializedRow := Row{}
	deserializeRow(serializedRow, &deserializedRow)

	if !reflect.DeepEqual(source, deserializedRow) {
		t.Errorf("Expected %v, but got %v.\n", source, deserializedRow)
	}
}
