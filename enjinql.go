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
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-corelibs/context"
	"github.com/go-corelibs/go-sqlbuilder"
)

// EnjinQL is the interface for a built enjinql instance
type EnjinQL interface {

	// Parse parses the Enjin Query Language format string and constructs a
	// new Syntax instance
	Parse(format string, args ...interface{}) (parsed *Syntax, err error)

	// ParsedToSql prepares the SQL query arguments from a parsed Syntax tree
	ParsedToSql(parsed *Syntax) (query string, argv []interface{}, err error)

	// ToSQL uses Parse and ParsedToSQL to produce the SQL query arguments
	ToSQL(format string, args ...interface{}) (query string, argv []interface{}, err error)

	// Perform uses ToSQL to build and execute the SQL statement
	Perform(format string, argv ...interface{}) (columns []string, results context.Contexts, err error)

	// Plan uses Parse to prepare the Syntax tree, then prepares the SQL table
	// INNER JOIN statement plan and returns two summaries of the resulting
	// plan: a brief one-liner and a verbose multi-line
	Plan(format string, args ...interface{}) (brief, verbose string, err error)

	// DBH returns either the current sql.Tx or the default sql.DB instance
	DBH() SqlDB

	// T returns the sqlbuilder.Table associated with the named source,
	// returns nil if the table or source do not exist
	T(name string) (t sqlbuilder.Table)

	// SqlBuilder returns a sqlbuilder.Buildable instance, preconfigured with
	// the EnjinQL sqlbuilder.Dialect
	SqlBuilder() sqlbuilder.Buildable

	// SqlDialect returns the configured go-sqlbuilder dialect
	SqlDialect() sqlbuilder.Dialect

	// SqlBegin starts and returns new SQL transaction, this is the only way
	// to properly add or remove data from indexing
	SqlBegin() (tx SqlTrunkTX, err error)

	// SqlExec is a convenience wrapper around sql.DB.Exec which returns the
	// sql.Result values in one step
	SqlExec(query string, argv ...interface{}) (id int64, affected int64, err error)

	// SqlQuery is a convenience wrapper around sql.DB.Query which returns
	// the column order and results
	SqlQuery(query string, argv ...interface{}) (columns []string, results context.Contexts, err error)

	// String returns an indented JSON representation of the Config
	String() string
	// Marshal returns a compact JSON representation of the Config
	Marshal() (data []byte, err error)
	// Unmarshal always returns the ErrUnmarshalEnjinQL error
	Unmarshal(data []byte) (err error)

	// Config returns a clone of this EnjinQL instance's configuration
	Config() (cloned *Config)

	// CreateTables will process all configured sources and issue CREATE TABLE
	// IF NOT EXISTS queries, stopping at the first error
	CreateTables() (err error)

	// CreateIndexes will process all configured sources and issue CREATE
	// INDEX IF NOT EXISTS queries, stopping at the first error
	CreateIndexes() (err error)

	// Close calls the Close method on the sql.DB instance and flags this
	// enjinql instance as being closed
	Close() (err error)

	// Ready returns nil if this EnjinQL instance has an open sql.DB instance
	// and returns sql.ErrConnDone otherwise
	Ready() error

	private(_ *enjinql) bool
}

type enjinql struct {
	closed bool
	option *option
	config *Config

	db      *cSqlDB
	dialect sqlbuilder.Dialect
	builder sqlbuilder.Buildable

	sources *cSources

	m *sync.RWMutex
}

type Option func(o *option) (err error)

type option struct {
	skipCreateTables  bool
	skipCreateIndexes bool
}

func SkipCreateTable(o *option) (err error) {
	o.skipCreateTables = true
	return
}

func SkipCreateIndex(o *option) (err error) {
	o.skipCreateIndexes = true
	return
}

func New(c *Config, dbh *sql.DB, dialect sqlbuilder.Dialect, options ...Option) (eql EnjinQL, err error) {
	if c == nil {
		err = fmt.Errorf("config is required")
		return
	} else if err = c.Validate(); err != nil {
		return
	} else if dbh == nil {
		err = fmt.Errorf("dbh is required")
		return
	} else if dialect == nil {
		err = fmt.Errorf("dialect is required")
		return
	}
	o := &option{}
	for _, ofn := range options {
		if err = ofn(o); err != nil {
			err = fmt.Errorf("enjinql.New option error: %w", err)
			return
		}
	}
	instance := &enjinql{
		option:  o,
		config:  c,
		dialect: dialect,
		builder: sqlbuilder.NewBuildable(dialect),
		m:       &sync.RWMutex{},
	}
	instance.db = newSqlDB(dbh, instance)
	if err = instance.init(); err == nil {
		eql = instance
	}
	return
}

func (eql *enjinql) init() (err error) {
	eql.sources = newSources(eql.config.Prefix, eql.builder)
	for _, sc := range eql.config.Sources {
		if err = eql.sources.addSource(sc); err != nil {
			err = fmt.Errorf("add source error: %w", err)
			return
		}
	}

	if !eql.option.skipCreateTables {
		if err = eql.CreateTables(); err != nil {
			return
		}
	}

	if !eql.option.skipCreateIndexes {
		if err = eql.CreateIndexes(); err != nil {
			return
		}
	}
	return
}

func (eql *enjinql) Close() (err error) {
	eql.m.Lock()
	defer eql.m.Unlock()
	if !eql.closed {
		eql.closed = true
		err = eql.db.db.Close()
	}
	return
}

func (eql *enjinql) Ready() error {
	eql.m.RLock()
	defer eql.m.RUnlock()
	if eql.closed {
		return sql.ErrConnDone
	}
	return nil
}

func (eql *enjinql) private(*enjinql) bool {
	// opsec measure to prevent false enjinql instances from being accepted as
	// real simply because they satisfy the exported methods in the EnjinQL
	// interface definition
	return true
}

func (eql *enjinql) CreateTables() (err error) {
	if err = eql.Ready(); err == nil {
		for _, sc := range eql.config.Sources {

			var query string
			var argv []interface{}
			var t sqlbuilder.Table

			var tx *sql.Tx
			if tx, err = eql.db.db.Begin(); err != nil {
				return
			}

			if source, ok := eql.sources.getSource(sc.Name); !ok {
				_ = tx.Rollback()
				err = fmt.Errorf("%w: %q", ErrSourceNotFound, sc.Name)
				return

			} else if t, err = source.getTable(); err != nil {
				_ = tx.Rollback()
				err = fmt.Errorf("%w: %q", ErrTableNotFound, source.formal())
				return

			} else if query, argv, err = eql.builder.CreateTable(t).IfNotExists().ToSql(); err != nil {
				_ = tx.Rollback()
				err = fmt.Errorf("%w: %q - %w", ErrCreateTableSQL, source.formal(), err)
				return

			} else if _, err = tx.Exec(query, argv...); err != nil {
				_ = tx.Rollback()
				err = fmt.Errorf("%w: %q - %w", ErrCreateTable, source.formal(), err)
				return

			} else if err = tx.Commit(); err != nil {
				_ = tx.Rollback()
				err = fmt.Errorf("error committing create table changes: %w", err)
				return

			}
		}
	}
	return
}

func (eql *enjinql) CreateIndexes() (err error) {
	if err = eql.Ready(); err == nil {
		for _, sc := range eql.config.Sources {

			var query string
			var argv []interface{}
			var t sqlbuilder.Table

			if source, ok := eql.sources.getSource(sc.Name); !ok {
				err = fmt.Errorf("%w: %q", ErrSourceNotFound, sc.Name)
				return

			} else if t, err = source.getTable(); err != nil {
				err = fmt.Errorf("%w: %q", ErrTableNotFound, source.formal())
				return

			} else {

				for _, index := range source.indexes {
					name := source.formal(index...)
					var columns []sqlbuilder.Column
					for _, key := range index {
						column := t.C(key)
						columns = append(columns, column)
					}

					var tx *sql.Tx
					if tx, err = eql.db.db.Begin(); err != nil {
						return
					}

					if query, argv, err = eql.builder.CreateIndex(t).Name(name).Columns(columns...).IfNotExists().ToSql(); err != nil {
						_ = tx.Rollback()
						// this is confirming go-sqlbuilder unit testing, no need to test again
						err = fmt.Errorf("%w: %q - %w", ErrCreateIndexSQL, name, err)
						return

					} else if _, err = eql.db.Exec(query, argv...); err != nil {
						_ = tx.Rollback()
						// this is confirming database/sql unit testing, no need to test again
						err = fmt.Errorf("%w: %q - %w", ErrCreateIndex, name, err)
						return

					} else if err = tx.Commit(); err != nil {
						_ = tx.Rollback()
						err = fmt.Errorf("error committing create table changes: %w", err)
						return

					}
				}

			}
		}
	}
	return
}

func (eql *enjinql) Config() (cloned *Config) {
	return eql.config.Clone()
}

func (eql *enjinql) Marshal() (data []byte, err error) {
	eql.m.RLock()
	defer eql.m.RUnlock()
	data, err = json.Marshal(eql.config)
	return
}

func (eql *enjinql) Unmarshal(_ []byte) (err error) {
	return ErrUnmarshalEnjinQL
}

func (eql *enjinql) String() string {
	eql.m.RLock()
	defer eql.m.RUnlock()
	return eql.config.String()
}

func (eql *enjinql) Parse(format string, args ...interface{}) (parsed *Syntax, err error) {
	eql.m.RLock()
	defer eql.m.RUnlock()
	var prepared string
	if prepared, err = PrepareSyntax(format, args...); err == nil && prepared != "" {
		parsed, err = ParseSyntax(prepared)
		return
	} else if err != nil {
		return
	}
	err = fmt.Errorf("%w: empty input", ErrInvalidSyntax)
	return
}

func (eql *enjinql) ParsedToSql(parsed *Syntax) (query string, argv []interface{}, err error) {
	eql.m.RLock()
	defer eql.m.RUnlock()
	query, argv, err = eql.prepareSQL(parsed)
	return
}

func (eql *enjinql) Plan(format string, args ...interface{}) (brief, verbose string, err error) {
	var parsed *Syntax
	var planned *gSourcePlan
	if parsed, err = eql.Parse(format, args...); err != nil {
		return
	} else if planned, err = eql.preparePlan(parsed); err != nil {
		return
	}
	brief = planned.String()
	verbose = planned.Verbose()
	return
}

func (eql *enjinql) ToSQL(format string, args ...interface{}) (query string, argv []interface{}, err error) {
	var parsed *Syntax
	if parsed, err = eql.Parse(format, args...); err != nil {
		return
	}
	query, argv, err = eql.ParsedToSql(parsed)
	return
}

func (eql *enjinql) Perform(format string, argv ...interface{}) (columns []string, results context.Contexts, err error) {
	if err = eql.Ready(); err == nil {
		var query string
		var args []interface{}
		if query, args, err = eql.ToSQL(format, argv...); err != nil {
			return
		}

		eql.m.RLock()
		defer eql.m.RUnlock()

		columns, results, err = eql.SqlQuery(query, args...)
	}
	return
}

func (eql *enjinql) DBH() SqlDB {
	return eql.db
}

func (eql *enjinql) T(name string) (t sqlbuilder.Table) {
	t, _ = eql.sources.T(name)
	return
}

func (eql *enjinql) SqlBuilder() sqlbuilder.Buildable {
	return eql.builder
}

func (eql *enjinql) SqlDialect() sqlbuilder.Dialect {
	return eql.dialect
}

func (eql *enjinql) SqlBegin() (tx SqlTrunkTX, err error) {
	return eql.db.begin(eql)
}

func (eql *enjinql) SqlExec(query string, argv ...interface{}) (id int64, affected int64, err error) {
	if err = eql.Ready(); err == nil {
		var result sql.Result
		if result, err = eql.db.Exec(query, argv...); err == nil {
			id, _ = result.LastInsertId()
			affected, _ = result.RowsAffected()
		}
	}
	return
}

func (eql *enjinql) SqlQuery(query string, argv ...interface{}) (columns []string, results context.Contexts, err error) {
	if err = eql.Ready(); err == nil {
		var rows *sql.Rows
		if rows, err = eql.db.Query(query, argv...); err == nil {

			for rows.Next() {
				var values []interface{}
				if len(columns) == 0 {
					columns, _ = rows.Columns()
				}
				for range columns {
					var v interface{} = nil
					values = append(values, &v)
				}
				_ = rows.Scan(values...) // safe to ignore because there are no scanner values
				row := context.New()
				for idx, name := range columns {
					if v, ok := values[idx].(*interface{}); ok {
						row[name] = *v
					}
				}
				results = append(results, row)
			}

		}
	}
	return
}
