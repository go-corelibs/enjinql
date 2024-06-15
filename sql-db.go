// Copyright (c) 2024  The Go-Enjin Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package enjinql

import (
	"context"
	"database/sql"

	clContext "github.com/go-corelibs/context"
)

var _ SqlDB = (*cSqlDB)(nil)

type SqlDB interface {
	Perform(format string, argv ...interface{}) (columns []string, results clContext.Contexts, err error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Prepare(query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Exec(query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryRow(query string, args ...any) *sql.Row
}

type cSqlDB struct {
	db  *sql.DB
	eql *enjinql
}

func newSqlDB(dbh *sql.DB, eql *enjinql) (c *cSqlDB) {
	return &cSqlDB{db: dbh, eql: eql}
}

func (c *cSqlDB) begin(eql *enjinql) (tx SqlTrunkTX, err error) {
	var transaction *sql.Tx
	if transaction, err = c.db.Begin(); err == nil {
		tx = newSqlTrunkTX(transaction, eql)
	}
	return
}

func (c *cSqlDB) Perform(format string, argv ...interface{}) (columns []string, results clContext.Contexts, err error) {
	columns, results, err = c.eql.Perform(format, argv...)
	return
}

func (c *cSqlDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return c.db.PrepareContext(ctx, query)
}

func (c *cSqlDB) Prepare(query string) (*sql.Stmt, error) {
	return c.db.Prepare(query)
}

func (c *cSqlDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

func (c *cSqlDB) Exec(query string, args ...any) (sql.Result, error) {
	return c.db.Exec(query, args...)
}

func (c *cSqlDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

func (c *cSqlDB) Query(query string, args ...any) (*sql.Rows, error) {
	return c.db.Query(query, args...)
}

func (c *cSqlDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return c.db.QueryRowContext(ctx, query, args...)
}

func (c *cSqlDB) QueryRow(query string, args ...any) *sql.Row {
	return c.db.QueryRow(query, args...)
}
