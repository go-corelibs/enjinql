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
)

var _ SqlTrunkTX = (*cSqlTrunkTX)(nil)

type SqlTrunkTX interface {
	SqlTX

	TX() SqlTX

	Valid() bool
	Commit() (err error)
	Rollback() (err error)
}

type cSqlTrunkTX struct {
	cSqlTX
}

func newSqlTrunkTX(tx *sql.Tx, eql *enjinql) *cSqlTrunkTX {
	return &cSqlTrunkTX{
		cSqlTX{
			tx:  tx,
			eql: eql,
		},
	}
}

func (c *cSqlTrunkTX) TX() SqlTX {
	return &cSqlTX{
		tx:  c.tx,
		eql: c.eql,
	}
}

func (c *cSqlTrunkTX) Valid() bool {
	return c.tx != nil
}

func (c *cSqlTrunkTX) Commit() (err error) {
	if c.Valid() {
		if err = c.tx.Commit(); err == nil {
			c.tx = nil
		}
	}
	return
}

func (c *cSqlTrunkTX) Rollback() (err error) {
	if c.Valid() {
		if err = c.tx.Rollback(); err == nil {
			c.tx = nil
		}
	}
	return
}
