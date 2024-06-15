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
	"fmt"

	clContext "github.com/go-corelibs/context"
	"github.com/go-corelibs/go-sqlbuilder"
)

var _ SqlTX = (*cSqlTX)(nil)

type SqlTX interface {
	SqlDB

	Insert(name string, values ...interface{}) (id int64, err error)
	Delete(name string, id int64) (affected int64, err error)
	DeleteWhereEQ(sourceName, key string, value interface{}) (affected int64, err error)
}

type cSqlTX struct {
	tx  *sql.Tx
	eql *enjinql
}

func (c *cSqlTX) Perform(format string, argv ...interface{}) (columns []string, results clContext.Contexts, err error) {
	columns, results, err = c.eql.Perform(format, argv...)
	return
}

func (c *cSqlTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return c.tx.PrepareContext(ctx, query)
}

func (c *cSqlTX) Prepare(query string) (*sql.Stmt, error) {
	return c.tx.Prepare(query)
}

func (c *cSqlTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return c.tx.ExecContext(ctx, query, args...)
}

func (c *cSqlTX) Exec(query string, args ...any) (sql.Result, error) {
	return c.tx.Exec(query, args...)
}

func (c *cSqlTX) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return c.tx.QueryContext(ctx, query, args...)
}

func (c *cSqlTX) Query(query string, args ...any) (*sql.Rows, error) {
	return c.tx.Query(query, args...)
}

func (c *cSqlTX) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return c.tx.QueryRowContext(ctx, query, args...)
}

func (c *cSqlTX) QueryRow(query string, args ...any) *sql.Row {
	return c.tx.QueryRow(query, args...)
}

func (c *cSqlTX) Insert(name string, values ...interface{}) (id int64, err error) {
	var ok bool
	var source *cSource
	if source, ok = c.eql.sources.getSource(name); !ok {
		err = fmt.Errorf("%w: %q", ErrSourceNotFound, name)
		return
	}
	table, _ := source.getTable()
	numValues := len(values)
	if numValues == 0 {
		err = fmt.Errorf("%w: %w", ErrInsertRow, ErrNoValues)
		return
	}

	var columns []sqlbuilder.Column
	for _, columnName := range source.order {
		columns = append(columns, table.C(columnName))
	}
	//columns = columns[1:] // drop the id column
	numColumns := len(columns)
	if numColumns < numValues {
		err = fmt.Errorf("%w: %w", ErrInsertRow, ErrTooManyValues)
		return
	} else if numColumns > numValues {
		columns = columns[:numValues]
	}

	b := c.eql.builder.Insert(table)
	b.Columns(columns...)
	b.Values(values...)

	var query string
	var argv []interface{}
	if query, argv, err = b.ToSql(); err != nil {
		// this specific error case can happen because of EnjinQL not
		// confirming the value types before this call
		err = fmt.Errorf("%w: %w", ErrInsertRow, err)
		return
	}

	var result sql.Result
	if result, err = c.tx.Exec(query, argv...); err != nil {
		// this is testing the go-sqlbuilder package and the underlying
		// database service
		err = fmt.Errorf("%w: %w", ErrInsertRow, err)
		return
	}

	id, err = result.LastInsertId()
	return
}

func (c *cSqlTX) Delete(name string, id int64) (affected int64, err error) {
	var ok bool
	var source *cSource
	if source, ok = c.eql.sources.getSource(name); !ok {
		err = fmt.Errorf("%w: %q", ErrSourceNotFound, name)
		return
	}
	table, _ := source.getTable()
	if id <= 0 {
		err = fmt.Errorf("%w: %w", ErrDeleteRows, ErrInvalidID)
		return
	}

	idColumn := table.C("id")
	b := c.eql.builder.
		Delete(table).
		Where(idColumn.Eq(id))

	var query string
	var argv []interface{}
	if query, argv, err = b.ToSql(); err != nil {
		// can this even happen? all the variables involved are confirmed so
		// checking for this specific error case is actually testing the
		// go-sqlbuilder package and not testing EnjinQL
		err = fmt.Errorf("%w: %w", ErrDeleteRows, err)
		return
	}

	var result sql.Result
	if result, err = c.tx.Exec(query, argv...); err != nil {
		// this is testing the go-sqlbuilder package and the underlying
		// database service
		err = fmt.Errorf("%w: %w", ErrDeleteRows, err)
		return
	}

	affected, err = result.RowsAffected()
	return
}

func (c *cSqlTX) DeleteWhereEQ(sourceName, key string, value interface{}) (affected int64, err error) {
	var ok bool
	var source *cSource
	if source, ok = c.eql.sources.getSource(sourceName); !ok {
		err = fmt.Errorf("%w: %q", ErrSourceNotFound, sourceName)
		return
	}
	table, _ := source.getTable()

	var column sqlbuilder.Column
	if column = table.C(key); column == nil {
		err = fmt.Errorf("%w: %q", ErrColumnNotFound, key)
		return
	}
	b := c.eql.builder.
		Delete(table).
		Where(column.Eq(value))

	var query string
	var argv []interface{}
	if query, argv, err = b.ToSql(); err != nil {
		// can this even happen? all the variables involved are confirmed so
		// checking for this specific error case is actually testing the
		// go-sqlbuilder package and not testing EnjinQL
		err = fmt.Errorf("%w: %w", ErrDeleteRows, err)
		return
	}

	var result sql.Result
	if result, err = c.tx.Exec(query, argv...); err != nil {
		// this is testing the go-sqlbuilder package and the underlying
		// database service
		err = fmt.Errorf("%w: %w", ErrDeleteRows, err)
		return
	}

	affected, err = result.RowsAffected()
	return
}
