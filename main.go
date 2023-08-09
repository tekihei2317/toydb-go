package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	buf.text = scanner.Text()
	buf.bufLen = len(buf.text)

	return nil
}

type MetaCommandResult int

const (
	META_COMMAND_SUCCESS MetaCommandResult = iota + 1
	META_COMMAND_UNRECOGNIZED_COMMAND
)

type PrepareResult int

const (
	PREPARE_SUCCESS PrepareResult = iota + 1
	PREPARE_UNRECOGNIZED_STATEMENT
)

type StatementType int

const (
	STATEMENT_INSERT StatementType = iota + 1
	STATEMENT_SELECT
)

type Statement struct {
	Type StatementType
}

func execMetaCommand(command string) MetaCommandResult {
	if command == ".exit" {
		os.Exit(0)
		return META_COMMAND_SUCCESS
	} else {
		return META_COMMAND_UNRECOGNIZED_COMMAND
	}
}

func prepareStatement(buf InputBuffer, statement *Statement) PrepareResult {
	if strings.HasPrefix(buf.text, "insert") {
		statement.Type = STATEMENT_INSERT
		return PREPARE_SUCCESS
	}
	if strings.HasPrefix(buf.text, "select") {
		statement.Type = STATEMENT_SELECT
		return PREPARE_SUCCESS
	}

	return PREPARE_UNRECOGNIZED_STATEMENT
}

func executeStatement(statement Statement) {
	switch statement.Type {
	case STATEMENT_INSERT:
		fmt.Printf("This is where we would do an insert.\n")
	case STATEMENT_SELECT:
		fmt.Printf("This is where we would do an select.\n")
	}

}

func main() {
	var buf InputBuffer
	for {
		printPrompt()
		readInput(&buf)

		if buf.text[0] == '.' {
			result := execMetaCommand(buf.text)

			if result == META_COMMAND_SUCCESS {
				continue
			} else {
				fmt.Printf("Unrecognized command '%s'.\n", buf.text)
				continue
			}
		}

		var statement Statement
		result := prepareStatement(buf, &statement)

		switch result {
		case PREPARE_SUCCESS:
		case PREPARE_UNRECOGNIZED_STATEMENT:
			fmt.Printf("Unrecognized keyword at start of '%s'.\n", buf.text)
			continue
		}

		executeStatement(statement)
		fmt.Printf("Executed\n")
	}
}
