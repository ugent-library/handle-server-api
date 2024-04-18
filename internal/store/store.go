package store

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	db *sql.DB
}

const (
	SQL_SELECT_ROW = "SELECT handle, idx, type, data, ttl, ttl_type, timestamp, admin_read, admin_write, pub_read, pub_write FROM handles WHERE handle = $1 AND type = 'URL' LIMIT 1"
	SQL_DELETE_ROW = "DELETE FROM handles WHERE handle = $1"
)

func connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
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

func (s *Store) Get(ctx context.Context, handle string) (*Handle, error) {
	var h Handle

	err := s.db.QueryRowContext(ctx, SQL_SELECT_ROW, handle).
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

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &h, nil
}

func (s *Store) Delete(ctx context.Context, handle string) (int64, error) {
	res, err := s.db.ExecContext(ctx, SQL_DELETE_ROW, handle)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (s *Store) Add(ctx context.Context, h *Handle) (int64, error) {
	sql := `
INSERT INTO handles(handle,idx,type,data,ttl,ttl_type,timestamp,admin_read,admin_write,pub_read,pub_write)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (handle, idx) DO UPDATE SET
idx = excluded.idx,
type = excluded.type,
data = excluded.data,
ttl = excluded.ttl,
ttl_type = excluded.ttl_type,
timestamp = excluded.timestamp,
admin_read = excluded.admin_read,
admin_write = excluded.admin_write,
pub_read = excluded.pub_read,
pub_write = excluded.pub_write
`

	res, err := s.db.ExecContext(
		ctx,
		sql,
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

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
