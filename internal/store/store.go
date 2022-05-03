package store

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type Store struct {
	db *sql.DB
}

const (
	SQL_SELECT_ROW = "SELECT handle, idx, type, data, ttl, ttl_type, timestamp, admin_read, admin_write, pub_read, pub_write FROM handles WHERE handle = ? AND type = 'URL' LIMIT 1"
)

func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(3)
	db.SetMaxOpenConns(3)

	return &Store{db: db}, nil
}

func (s *Store) Get(handle string) *Handle {

	var h Handle

	err := s.db.QueryRow(SQL_SELECT_ROW, handle).
		Scan(
			&h.Handle,
			&h.Idx,
			&h.Type,
			&h.Data,
			&h.Ttl,
			&h.TtlType,
			&h.Timestamp,
			&h.AdminRead,
			&h.AdminWrite,
			&h.PubRead,
			&h.PubWrite)

	if err != nil {

		if err == sql.ErrNoRows {
			return nil
		}
		log.Fatal(err.Error())

	}

	return &h
}

func (s *Store) Delete(handle string) int64 {

	res, err := s.db.Exec("DELETE FROM handles WHERE handle = ?", handle)

	if err != nil {

		log.Fatal(err.Error())

	}

	rowsAffected, _ := res.RowsAffected()

	return rowsAffected
}
