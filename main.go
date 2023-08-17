package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"toydb-go/core"
	"toydb-go/execute"
	db "toydb-go/table"
)

type InputBuffer struct {
	text     string
	bufLen   int
	inputLen int
}

func printPrompt() {
	fmt.Print("db > ")
}

func readInput(scanner *bufio.Scanner, buf *InputBuffer) error {
	scanner.Scan()
	buf.text = scanner.Text()
	buf.bufLen = len(buf.text)

	return nil
}

type PrepareResult int

const (
	PREPARE_SUCCESS PrepareResult = iota + 1
	PREPARE_UNRECOGNIZED_STATEMENT
	PREPARE_SYNTAX_ERROR
	PREPARE_STRING_TOO_LONG
	PREPARE_NEGATIVE_ID
)

// 入力からステートメントを作成する
func prepareStatement(buf InputBuffer, statement *core.Statement) PrepareResult {
	if strings.HasPrefix(buf.text, "insert") {
		statement.Type = core.STATEMENT_INSERT

		var id int
		var username, email string
		assigned, _ := fmt.Sscanf(buf.text, "insert %d %s %s", &id, &username, &email)

		if assigned != 3 {
			return PREPARE_SYNTAX_ERROR
		}

		if id < 0 {
			return PREPARE_NEGATIVE_ID
		}

		if len(username) > db.USERNAME_SIZE {
			return PREPARE_STRING_TOO_LONG
		}

		if len(email) > db.EMAIL_SIZE {
			return PREPARE_STRING_TOO_LONG
		}

		statement.RowToInsert.Id = id
		copy(statement.RowToInsert.Username[:], username)
		copy(statement.RowToInsert.Email[:], email)

		return PREPARE_SUCCESS
	}
	if strings.HasPrefix(buf.text, "select") {
		statement.Type = core.STATEMENT_SELECT
		return PREPARE_SUCCESS
	}

	return PREPARE_UNRECOGNIZED_STATEMENT
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Must supply a database filename.")
		os.Exit(1)
	}

	table, err := db.DbOpen(os.Args[1])
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	var buf InputBuffer

	for {
		printPrompt()
		readInput(scanner, &buf)

		if buf.text[0] == '.' {
			result := execute.ExecMetaCommand(buf.text, table)

			if result == execute.META_COMMAND_SUCCESS {
				continue
			} else {
				fmt.Printf("Unrecognized command '%s'.\n", buf.text)
				continue
			}
		}

		var statement core.Statement
		result := prepareStatement(buf, &statement)

		switch result {
		case PREPARE_SUCCESS:
		case PREPARE_SYNTAX_ERROR:
			fmt.Printf("Syntax error. Could not parse statement.\n")
			continue
		case PREPARE_UNRECOGNIZED_STATEMENT:
			fmt.Printf("Unrecognized keyword at start of '%s'.\n", buf.text)
			continue
		case PREPARE_STRING_TOO_LONG:
			fmt.Printf("String is too long.\n")
			continue
		case PREPARE_NEGATIVE_ID:
			fmt.Printf("ID must be positive.\n")
			continue
		}

		executeResult := execute.ExecuteStatement(statement, table)
		switch executeResult {
		case execute.EXECUTE_SUCCESS:
			fmt.Printf("Executed.\n")
		case execute.EXECUTE_TABLE_FULL:
			fmt.Printf("Error: Table is full.\n")
		}
	}
}
