package database

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type DB struct {
	*sql.DB
}

func New(params Params) (*DB, error) {
	psqlconn := params.ConnStr()
	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return nil, errors.Wrap(err, "sql.Open")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "db.Ping")
	}

	return &DB{db}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}
