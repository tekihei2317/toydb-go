package main

import "fmt"

// やりたいこと
// user (id, name, email)をシリアライズして、メモリに保存する。メモリは動的に確保する。
// メモリからUserの構造体にデシリアライズする。

func stringTest() {
	str := "あいab"

	// バイトでアクセスする
	for i := 0; i < len(str); i++ {
		fmt.Printf("%x ", str[i])
	}
	fmt.Println()

	// 文字列ごとにアクセスする
	for i, r := range str {
		fmt.Printf("%d %c\n", i, r)
	}

	// lenはバイト数
	fmt.Printf("len(str): %d\n", len(str))
	fmt.Printf("len(cast str as rune[]): %d\n", len(([]rune)(str)))
}

// func bytesToString(bytes []byte) string {
// 	var validBytes []byte
// 	for _, b := range bytes {
// 		if b == 0 {
// 			break
// 		}
// 		validBytes = append(validBytes, b)
// 	}
// 	return string(validBytes)
// }

// func bytesToStringTest() {
// 	bytes := [32]byte{'a', 'i'}
// 	str := string(bytes[:])

// 	fmt.Printf("%s\n", str)
// 	// nilを出力すると、ヌル文字が入るので除外する
// 	fmt.Printf("%s\n", bytesToString(bytes[:]))
// }
