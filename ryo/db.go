package ryo

import (
	"database/sql"
	"database/sql/driver"
	"log"
	"time"
)

type DB struct {
	Underlying *sql.DB
}

func Open(driverName, dataSourceName string) (*DB, error) {
	sqlDB, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		if sqlDB != nil {
			sqlDB.Close()
		}
		return nil, err
	}

	return &DB{sqlDB}, nil
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.Underlying.Begin()
	if err != nil {
		if tx != nil {
			defer tx.Rollback()
		}
		return nil, err
	}

	return &Tx{tx}, nil
}

func (db *DB) Close() error {
	return db.Underlying.Close()
}

func (db *DB) Driver() driver.Driver {
	return db.Underlying.Driver()
}

func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	log.Println(query)
	for i, v := range args {
		log.Printf("  $%d => %v", i+1, v)
	}
	return db.Underlying.Exec(query, args...)
}

func (db *DB) Ping() error {
	return db.Underlying.Ping()
}

func (db *DB) Prepare(query string) (*sql.Stmt, error) {
	return db.Underlying.Prepare(query)
}

func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.Underlying.Query(query, args...)
}

func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.Underlying.QueryRow(query, args...)
}

func (db *DB) SetConnMaxLifetime(d time.Duration) {
	db.Underlying.SetConnMaxLifetime(d)
}

func (db *DB) SetMaxIdleConns(n int) {
	db.Underlying.SetMaxIdleConns(n)
}

func (db *DB) SetMaxOpenConns(n int) {
	db.Underlying.SetMaxOpenConns(n)
}

func (db *DB) Stats() sql.DBStats {
	return db.Underlying.Stats()
}
