package main

import (
	"fmt"
	"testing"
)

func TestTable(t *testing.T) {
	var table Table

	fmt.Printf("%+v", table.pages)
}
