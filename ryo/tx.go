package ryo

import (
	"database/sql"
)

type Tx struct {
	Underlying *sql.Tx
}

func (tx *Tx) Commit() error {
	return tx.Underlying.Commit()
}

func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.Underlying.Exec(query, args...)
}

func (tx *Tx) Prepare(query string) (*sql.Stmt, error) {
	return tx.Underlying.Prepare(query)
}

func (tx *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tx.Underlying.Query(query, args...)
}

func (tx *Tx) QueryRow(query string, args ...interface{}) *sql.Row {
	return tx.Underlying.QueryRow(query, args...)
}

func (tx *Tx) Rollback() error {
	return tx.Underlying.Rollback()
}

func (tx *Tx) Stmt(stmt *sql.Stmt) *sql.Stmt {
	return tx.Underlying.Stmt(stmt)
}
