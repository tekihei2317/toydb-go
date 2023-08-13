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
