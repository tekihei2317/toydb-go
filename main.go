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

func readInput(scanner *bufio.Scanner, buf *InputBuffer) error {
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
	PREPARE_SYNTAX_ERROR
	PREPARE_STRING_TOO_LONG
	PREPARE_NEGATIVE_ID
)

type StatementType int

const (
	STATEMENT_INSERT StatementType = iota + 1
	STATEMENT_SELECT
)

type Row struct {
	id       int
	username [32]byte
	email    [256]byte
}

type Statement struct {
	Type        StatementType
	RowToInsert Row // insert only
}

func execMetaCommand(command string, table *Table) MetaCommandResult {
	if command == ".exit" {
		dbClose(table)
		os.Exit(0)
		return META_COMMAND_SUCCESS
	} else {
		return META_COMMAND_UNRECOGNIZED_COMMAND
	}
}

func prepareStatement(buf InputBuffer, statement *Statement) PrepareResult {
	if strings.HasPrefix(buf.text, "insert") {
		statement.Type = STATEMENT_INSERT

		var id int
		var username, email string
		assigned, _ := fmt.Sscanf(buf.text, "insert %d %s %s", &id, &username, &email)

		if assigned != 3 {
			return PREPARE_SYNTAX_ERROR
		}

		if id < 0 {
			return PREPARE_NEGATIVE_ID
		}

		if len(username) > USERNAME_SIZE {
			return PREPARE_STRING_TOO_LONG
		}

		if len(email) > EMAIL_SIZE {
			return PREPARE_STRING_TOO_LONG
		}

		statement.RowToInsert.id = id
		copy(statement.RowToInsert.username[:], username)
		copy(statement.RowToInsert.email[:], email)

		return PREPARE_SUCCESS
	}
	if strings.HasPrefix(buf.text, "select") {
		statement.Type = STATEMENT_SELECT
		return PREPARE_SUCCESS
	}

	return PREPARE_UNRECOGNIZED_STATEMENT
}

type ExecuteResult int

const (
	EXECUTE_SUCCESS ExecuteResult = iota + 1
	EXECUTE_TABLE_FULL
)

func bytesToString(bytes []byte) string {
	var validBytes []byte
	for _, b := range bytes {
		if b == 0 {
			break
		}
		validBytes = append(validBytes, b)
	}
	return string(validBytes)
}

func printRow(row *Row) {
	fmt.Printf("(%d, %s, %s)\n", row.id, bytesToString(row.username[:]), bytesToString(row.email[:]))
}

func executeSelect(statement Statement, table *Table) ExecuteResult {
	cursor := tableStart(table)

	for !cursor.endOfTable {
		row := table.getRowByRowNum(cursor.rowNum)
		printRow(&row)
		cursorAdvance(&cursor)
	}

	return EXECUTE_SUCCESS
}

func executeInsert(statement Statement, table *Table) ExecuteResult {
	if table.numRows >= uint32(TABLE_MAX_ROWS) {
		return EXECUTE_TABLE_FULL
	}

	rowToInsert := &statement.RowToInsert
	table.insertRow(rowToInsert)

	return EXECUTE_SUCCESS
}

func executeStatement(statement Statement, table *Table) ExecuteResult {
	switch statement.Type {
	case STATEMENT_INSERT:
		return executeInsert(statement, table)
	case STATEMENT_SELECT:
		return executeSelect(statement, table)
	default:
		return EXECUTE_SUCCESS
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Must supply a database filename.")
		os.Exit(1)
	}

	table, err := dbOpen(os.Args[1])
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
			result := execMetaCommand(buf.text, table)

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

		executeResult := executeStatement(statement, table)
		switch executeResult {
		case EXECUTE_SUCCESS:
			fmt.Printf("Executed.\n")
		case EXECUTE_TABLE_FULL:
			fmt.Printf("Error: Table is full.\n")
		}
	}
}
