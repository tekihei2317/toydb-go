package main

import (
	"fmt"
	"os"
)

type InputBuffer struct {
	text     string
	bufLen   int
	inputLen int
}

func printPrompt() {
	fmt.Print("db > ")
}

func readInput(buf *InputBuffer) error {
	_, err := fmt.Scanln(&buf.text)
	buf.bufLen = len(buf.text)

	if err != nil {
		return err
	}
	return nil
}

func main() {
	var buf InputBuffer
	for {
		printPrompt()
		readInput(&buf)

		if buf.text == ".exit" {
			os.Exit(0)
		} else {
			fmt.Printf("Unrecognized command '%s'\n", buf.text)
		}
	}
}
