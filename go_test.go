package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
	"testing"
	"unsafe"
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

type User struct {
	id     int
	emails []string
}

func modifyUserEmail(user *User) *User {
	// user.emails = []string{"updated@example.com"}
	user.emails[0] = "updated@example.com"

	return user
}

func TestSliceInStruct(t *testing.T) {
	// 構造体の中にスライスがある場合、ポインタ渡しでコピーされるのか
	user := User{
		id:     1,
		emails: []string{"tekihei@example.com"},
	}

	modifyUserEmail(&user)

	// 関数内でのスライスの書き換えは、変更元にも反映される（構造体をポイント渡ししているのでそれはそう？）
	if !reflect.DeepEqual(user.emails, []string{"updated@example.com"}) {
		t.Errorf("user.emails is %v\n", user.emails)
	}
}

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

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func TestReadFile(t *testing.T) {
	f, err := os.Open("./README.md")
	check(err)

	// 1. 1行だけ読み取る
	b1 := make([]byte, 10)
	// f.Readは、スライスの数だけファイルを読み込む
	f.Read(b1)

	expected := "# toydb-go"
	if !reflect.DeepEqual(b1, []byte(expected)) {
		t.Errorf("expected %v, but got %v", []byte(expected), b1)
	}

	// 2. 途中から読み取る
	f2, err := os.Open("./README.md")
	check(err)
	_, err2 := f2.Seek(2, 0)
	check(err2)

	b2 := make([]byte, 8)
	f2.Read(b2)

	expected2 := "toydb-go"
	if !reflect.DeepEqual(b2, []byte(expected2)) {
		t.Errorf("expected %v, but got %v", []byte(expected2), b2)
	}

	// 2.1 途中から読み取る（File.ReadAt）
	f3, err := os.Open("./README.md")
	check(err)

	b3 := make([]byte, 8)
	_, err3 := f3.ReadAt(b3, 2)
	check(err3)

	expected3 := "toydb-go"
	if !reflect.DeepEqual(b3, []byte(expected3)) {
		t.Errorf("expected %v, but got %v", []byte(expected3), b3)
	}
}

func TestWriteFile(t *testing.T) {
	// なければ作成し、あればそのまま読み取る
	// f, err := os.OpenFile("file", os.O_RDWR|os.O_CREATE, 0666)

	// 1. 先頭から書き込む
	f, err := os.Create("file") // なければ作成し、あれば空っぽにする
	check(err)
	defer f.Close()

	f.Write([]byte("abcde"))

	content, err := os.ReadFile("file")
	check(err)

	expected := []byte("abcde")
	if !reflect.DeepEqual(content, expected) {
		t.Errorf("expected %v, but got %v", expected, content)
	}

	// 2. 途中から書き込む
	offset, err := f.Seek(2, 0)
	if offset != 2 {
		t.Errorf("offset is %d", offset)
	}
	f.Write([]byte("aaaaa"))

	content2, err := os.ReadFile("file")
	check(err)

	expected2 := []byte("abaaaaa")
	if !reflect.DeepEqual(content2, expected2) {
		t.Errorf("expected %v, but got %v", expected2, content2)
	}
}

func TestByteToInt(t *testing.T) {
	type Row struct {
		Id       int
		Username [32]byte
	}

	// 構造体→[]byte
	row := Row{Id: 1, Username: [32]byte{'u', 's', 'e', 'r'}}
	src := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(&row))
	dst := make([]byte, unsafe.Sizeof(Row{}))
	copy(dst, src[:])

	if !reflect.DeepEqual(dst, src[:]) {
		t.Errorf("dst and src is not equal")
	}
	fmt.Println(src, dst)

	// []byte→構造体
	deserializedRow := &Row{}
	dst2 := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(deserializedRow))
	copy(dst2[:], dst)

	if !reflect.DeepEqual(deserializedRow.Username, [32]byte{'u', 's', 'e', 'r'}) {
		t.Errorf("Username is not equal")
	}

	// []byteの一部→int
	idBytes := dst[0:8]
	num := (*int)(unsafe.Pointer(&idBytes))
	num2 := binary.LittleEndian.Uint64(idBytes)

	fmt.Println(*num)
	fmt.Println(num2)

	numBytes := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(numBytes, 100)
	num3 := binary.LittleEndian.Uint64(numBytes)
	fmt.Println(num3)
}
