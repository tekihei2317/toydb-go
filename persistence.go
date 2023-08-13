package main

import (
	"fmt"
	"unsafe"
)

const (
	ID_OFFSET       = 0
	USERNAME_OFFSET = ID_OFFSET + int(unsafe.Sizeof(int(0)))
	EMAIL_OFFSET    = USERNAME_OFFSET + int(unsafe.Sizeof([32]byte{}))
	ID_SIZE         = int(unsafe.Sizeof(int(0)))
	USERNAME_SIZE   = int(unsafe.Sizeof([32]byte{}))
	EMAIL_SIZE      = int(unsafe.Sizeof([256]byte{}))
	ROW_SIZE        = ID_SIZE + USERNAME_SIZE + EMAIL_SIZE
)

// 行の構造体を、バイトのスライスに書き込む
func serializeRow(source *Row, destination []byte) {
	sourceBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(source))

	fmt.Println("copy: ", sourceBytes[:])
	fmt.Println("copy into: ", destination[0:ROW_SIZE])

	copy(destination[0:ROW_SIZE], sourceBytes[:])

	fmt.Println("copied:", destination[0:ROW_SIZE])
}

// 行データのスライスを、行の構造体に書き込む
func deserializeRow(source []byte, destination *Row) {
	destinationBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(destination))

	copy(destinationBytes[:], source[0:ROW_SIZE])
}
