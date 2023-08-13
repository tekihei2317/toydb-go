package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSscanf(t *testing.T) {
	var num1, num2 int
	var str1, str2 string
	c, _ := fmt.Sscanf("select 1 str", "select %d %s", &num1, &str1)
	c2, _ := fmt.Sscanf("select 1", "select %d %s", &num2, &str2)

	if c != 2 || num1 != 1 || str1 != "str" {
		t.Error("Parse failed.")
	}

	// パースに失敗してもエラーにはならない。変数にはゼロ値が入ったままになる
	if c2 != 1 || num2 != 1 || str2 != "" {
		t.Error("Parse failed.")
	}
}

func TestStringToByteSlice(t *testing.T) {
	str := "abc"
	str2 := "あいう"

	expected := []byte{97, 98, 99}
	if !reflect.DeepEqual([]byte(str), expected) {
		t.Errorf("expected %+v, got %+v\n", expected, []byte(str))
	}

	expected2 := []byte{
		227,
		129,
		130,
		227,
		129,
		132,
		227,
		129,
		134,
	}
	if !reflect.DeepEqual([]byte(str2), expected2) {
		t.Errorf("expected %+v, got %+v\n", expected2, []byte(str2))
	}
}

func TestWriteBytes(t *testing.T) {
	// スライスの内容を別のスライスに書き込むためには、copyを使用する

	var dest []byte

	bytes := []byte("abcde")
	copy(dest, bytes)

	// 元のスライスの容量だけコピーされる。srcが空のスライスだったら空のまま。
	// 空のスライスっぽかったけどテストが通らなかった
	// if !reflect.DeepEqual(dest, []byte{}) {
	// 	t.Errorf("expected: %+v, actual: %+v", []byte{}, dest)
	// }

	// 先にスライスの容量を確保してからコピーする
	dest = make([]byte, len(bytes))
	copy(dest, bytes)

	if !reflect.DeepEqual(dest, bytes) {
		t.Errorf("expected: %+v, actual: %+v", bytes, dest)
	}
}

func TestRestoreBytes(t *testing.T) {
	var dest []byte

	// 文字列をdestに2つ書き込んで、復元する
	s1 := "abc"
	s2 := "あいうえお"

	// 2つの文字列のバイト数だけ確保する
	dest = make([]byte, len(s1)+len(s2))
	copy(dest, []byte(s1))
	copy(dest[len(s1):], []byte(s2))

	// 復元する（コピー先のスライスの、開始位置・終了位置を指定できる）
	var restored = make([]byte, len(s1)+len(s2))
	copy(restored[0:], dest[0:len(s1)])
	copy(restored[len(s1):], dest[len(s1):len(s1)+len(s2)])

	if !reflect.DeepEqual(dest, restored) {
		t.Errorf("dest and restored doesn't match. dest: %+v, restored: %+v", dest, restored)
	}
}

func TestConvertStructToBytes(t *testing.T) {
	type User struct {
		id   int
		name [16]byte
	}

	// user := User{
	// 	id:   1,
	// 	name: [16]byte{'t', 'e', 'k', 'i', 'h', 'e', 'i'},
	// }

	// これでいいのか...？
	// userBytes := (*[unsafe.Sizeof(User{})]byte)(unsafe.Pointer(&user))
	// fmt.Println(userBytes)
}

func modifySlice(a []int) {
	a[2] = 100
}

func appendSlice(a []int) []int {
	return append(a, 100)
}

func copySlice(a []int, b []int) []int {
	copy(a, b)

	return a
}

func TestModifySlice(t *testing.T) {
	var a []int
	for i := 0; i < 5; i++ {
		a = append(a, i)
	}

	modifySlice(a)
	// スライスを関数に渡して要素を変更すると、元のデータも変わる
	if !reflect.DeepEqual(a, []int{0, 1, 100, 3, 4}) {
		t.Errorf("a is %v\n", a)
	}

	var b []int
	for i := 0; i < 5; i++ {
		b = append(b, i)
	}

	appended := appendSlice(b)
	if !reflect.DeepEqual(appended, []int{0, 1, 2, 3, 4, 100}) {
		t.Errorf("appended is %v\n", appended)
	}

	// スライスを関数に渡して要素を追加しても、元のデータは変わらない
	// 関数の側で、要素が不足していて新しいスライスが作成されるため
	if !reflect.DeepEqual(b, []int{0, 1, 2, 3, 4}) {
		t.Errorf("b is %v\n", b)
	}
}

func TestCopySlice(t *testing.T) {
	// 関数に渡してスライスをコピーしたときに、元の側に反映されるか確認する
	var a = []int{0, 1, 2, 3}
	var b = []int{100, 100}

	copied := copySlice(a, b)

	if !reflect.DeepEqual(a, []int{100, 100, 2, 3}) {
		t.Errorf("a is %v\n", a)
	}
	if !reflect.DeepEqual(copied, []int{100, 100, 2, 3}) {
		t.Errorf("copied is %v\n", copied)
	}

	var c = []int{100, 100}
	var d = []int{0, 1, 2, 3}

	copied2 := copySlice(c, d)

	// 容量を超える分はコピーされないので、copy(A, B)では新しいスライスが作成されることはなさそう
	if !reflect.DeepEqual(c, []int{0, 1}) {
		t.Errorf("c is %v\n", c)
	}
	if !reflect.DeepEqual(copied2, []int{0, 1}) {
		t.Errorf("copied2 is %v\n", copied2)
	}
}
