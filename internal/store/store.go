package store

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type Store struct {
	db *sql.DB
}

const (
	SQL_SELECT_ROW = "SELECT handle, idx, type, data, ttl, ttl_type, timestamp, admin_read, admin_write, pub_read, pub_write FROM handles WHERE handle = ? AND type = 'URL' LIMIT 1"
	SQL_DELETE     = "DELETE FROM handles WHERE handle = ?"
)

func connect(dsn string) (*sql.DB, error) {

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)

	return db, nil
}

func NewStore(dsn string) (*Store, error) {
	db, err := connect(dsn)

	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Get(handle string) (*Handle, error) {

	if e := s.db.Ping(); e != nil {
		return nil, e
	}

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
			return nil, nil
		}
		return nil, err

	}

	return &h, nil
}

func (s *Store) Delete(handle string) (int64, error) {

	if e := s.db.Ping(); e != nil {
		return 0, e
	}

	res, err := s.db.Exec(SQL_DELETE, handle)

	if err != nil {

		return 0, err

	}

	rowsAffected, _ := res.RowsAffected()

	return rowsAffected, nil
}

func (s *Store) Add(h *Handle) (int64, error) {

	if e := s.db.Ping(); e != nil {
		return 0, e
	}

	//cf. https://dev.mysql.com/doc/refman/8.0/en/replace.html
	res, err := s.db.Exec(
		"REPLACE INTO handles(handle,idx,type,data,ttl,ttl_type,timestamp,admin_read,admin_write,pub_read,pub_write) VALUES(?,?,?,?,?,?,?,?,?,?,?)",
		h.Handle,
		h.Idx,
		h.Type,
		h.Data,
		h.Ttl,
		h.TtlType,
		h.Timestamp,
		h.AdminRead,
		h.AdminWrite,
		h.PubRead,
		h.PubWrite,
	)

	if err != nil {

		return 0, err

	}

	rowsAffected, _ := res.RowsAffected()

	return rowsAffected, nil
}
