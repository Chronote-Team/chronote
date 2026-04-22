package db

import (
	"context"
	"database/sql"
)

type Checker struct {
	db *sql.DB
}

func NewChecker(db *sql.DB) *Checker {
	return &Checker{db: db}
}

func (c *Checker) Check(ctx context.Context) error {
	if c == nil || c.db == nil {
		return sql.ErrConnDone
	}
	return c.db.PingContext(ctx)
}
