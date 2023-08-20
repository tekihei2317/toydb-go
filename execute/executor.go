package execute

import (
	"fmt"
	"os"
	"toydb-go/core"
	db "toydb-go/table"
)

type ExecuteResult int

type MetaCommandResult int

const (
	META_COMMAND_SUCCESS MetaCommandResult = iota + 1
	META_COMMAND_UNRECOGNIZED_COMMAND
)

const (
	EXECUTE_SUCCESS ExecuteResult = iota + 1
	EXECUTE_TABLE_FULL
	EXECUTE_DUPLICATE_KEY
)

// メタコマンドを実行する
func ExecMetaCommand(command string, table *db.Table) MetaCommandResult {
	if command == ".exit" {
		db.DbClose(table)
		os.Exit(0)
		return META_COMMAND_SUCCESS
	} else if command == ".btree" {
		fmt.Println("Tree:")
		db.PrintTree(table, 0, 0)
		return META_COMMAND_SUCCESS
	} else {
		return META_COMMAND_UNRECOGNIZED_COMMAND
	}
}

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

func printRow(row *db.Row) {
	fmt.Printf("(%d, %s, %s)\n", row.Id, bytesToString(row.Username[:]), bytesToString(row.Email[:]))
}

// SELECT文を実行する
func executeSelect(statement core.Statement, table *db.Table) ExecuteResult {
	cursor := db.TableStart(table)

	for !cursor.EndOfTable {
		row := table.GetRowByCursor(cursor.PageNum, cursor.CellNum)
		printRow(&row)
		db.CursorAdvance(cursor)
	}

	return EXECUTE_SUCCESS
}

// INSERT文を実行する
func executeInsert(statement core.Statement, table *db.Table) ExecuteResult {
	rowToInsert := &statement.RowToInsert
	insertResult := table.InsertRow(&db.Row{
		Id:       rowToInsert.Id,
		Username: rowToInsert.Username,
		Email:    rowToInsert.Email,
	})

	switch insertResult {
	case db.INSERT_SUCCESS:
		return EXECUTE_SUCCESS
	case db.INSERT_DUPLICATE_KEY:
		return EXECUTE_DUPLICATE_KEY
	case db.INSERT_TABLE_FULL:
		return EXECUTE_TABLE_FULL
	default:
		// ここを通らないことをアサーションしたい、panicでいいのかな
		return EXECUTE_SUCCESS
	}
}

// SQLステートメントを実行する
func ExecuteStatement(statement core.Statement, table *db.Table) ExecuteResult {
	switch statement.Type {
	case core.STATEMENT_INSERT:
		return executeInsert(statement, table)
	case core.STATEMENT_SELECT:
		return executeSelect(statement, table)
	default:
		return EXECUTE_SUCCESS
	}
}
