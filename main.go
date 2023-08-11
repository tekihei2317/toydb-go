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
	PREPARE_SYNTAX_ERROR
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

		var username, email string
		assigned, _ := fmt.Sscanf(buf.text, "insert %d %s %s", &statement.RowToInsert.id, &username, &email)
		copy(statement.RowToInsert.username[:], username)
		copy(statement.RowToInsert.email[:], email)

		if assigned != 3 {
			return PREPARE_SYNTAX_ERROR
		}

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

func printRow(row *Row) {
	fmt.Printf("(%d, %s, %s)\n", row.id, row.username, row.email)
}

func executeSelect(statement Statement, table *Table) ExecuteResult {
	var row Row
	for i := uint32(0); i < table.numRows; i++ {
		deserializeRow(rowSlot(table, i), &row)
		printRow(&row)
	}
	return EXECUTE_SUCCESS
}

func executeInsert(statement Statement, table *Table) ExecuteResult {
	if table.numRows >= TABLE_MAX_ROWS {
		return EXECUTE_TABLE_FULL
	}

	rowToInsert := &statement.RowToInsert
	serializeRow(rowToInsert, rowSlot(table, table.numRows))
	table.numRows += 1

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
	table := Table{numRows: 0}
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
		case PREPARE_SYNTAX_ERROR:
			fmt.Printf("Syntax error. Could not parse statement.\n")
			continue
		case PREPARE_UNRECOGNIZED_STATEMENT:
			fmt.Printf("Unrecognized keyword at start of '%s'.\n", buf.text)
			continue
		}

		executeResult := executeStatement(statement, &table)
		switch executeResult {
		case EXECUTE_SUCCESS:
			fmt.Printf("Executed\n")
		case EXECUTE_TABLE_FULL:
			fmt.Printf("Error: Table is full.\n")
		}
	}
}
