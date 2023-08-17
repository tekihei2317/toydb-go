package core

type Row struct {
	Id       int
	Username [32]byte
	Email    [256]byte
}

type StatementType int

const (
	STATEMENT_INSERT StatementType = iota + 1
	STATEMENT_SELECT
)

type Statement struct {
	Type        StatementType
	RowToInsert Row // insert only
}
