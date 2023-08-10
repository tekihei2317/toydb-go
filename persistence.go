package main

import "unsafe"

const (
	ID_OFFSET       = 0
	USERNAME_OFFSET = ID_OFFSET + int(unsafe.Sizeof(int(0)))
	EMAIL_OFFSET    = USERNAME_OFFSET + int(unsafe.Sizeof([32]byte{}))
	ID_SIZE         = int(unsafe.Sizeof(int(0)))
	USERNAME_SIZE   = int(unsafe.Sizeof([32]byte{}))
	EMAIL_SIZE      = int(unsafe.Sizeof([256]byte{}))
	ROW_SIZE        = ID_SIZE + USERNAME_SIZE + EMAIL_SIZE
)

func serializeRow(source *Row, destination unsafe.Pointer) {
	destinationBytes := (*[unsafe.Sizeof(Row{})]byte)(destination)
	sourceBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(source))

	copy(destinationBytes[ID_OFFSET:], sourceBytes[ID_OFFSET:])
	copy(destinationBytes[USERNAME_OFFSET:], sourceBytes[USERNAME_OFFSET:])
	copy(destinationBytes[EMAIL_OFFSET:], sourceBytes[EMAIL_OFFSET:])
}

func deserializeRow(source unsafe.Pointer, destination *Row) {
	sourceBytes := (*[unsafe.Sizeof(Row{})]byte)(source)
	destinationBytes := (*[unsafe.Sizeof(Row{})]byte)(unsafe.Pointer(destination))

	copy(destinationBytes[ID_OFFSET:], sourceBytes[ID_OFFSET:])
	copy(destinationBytes[USERNAME_OFFSET:], sourceBytes[USERNAME_OFFSET:])
	copy(destinationBytes[EMAIL_OFFSET:], sourceBytes[EMAIL_OFFSET:])
}
