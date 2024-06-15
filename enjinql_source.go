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
	"fmt"

	"github.com/iancoleman/strcase"

	"github.com/go-corelibs/go-sqlbuilder"
	"github.com/go-corelibs/maps"
	"github.com/go-corelibs/slices"
)

const (
	// SourceIdKey is the SQL name of the primary key for all SQL tables
	SourceIdKey = "id"
)

type cSource struct {
	name    string
	node    *gSourceNode
	idxs    *cSources
	parent  *cSource
	keys    map[string]sqlbuilder.Column
	order   []string
	value   cSourceValue
	values  []cSourceValue
	unique  [][]string
	indexes [][]string
	table   sqlbuilder.Table
	column  map[string]sqlbuilder.ColumnConfig
	links   map[string]string
}

func (c *cSource) init() (err error) {

	// validate the options

	switch {
	case c.value.ivt == gInvalidValue:
		err = fmt.Errorf("invalid source value type for: %q", c.value.key)
		return
	}

	for _, v := range c.values {
		if v.ivt == gInvalidValue {
			err = fmt.Errorf("invalid source value type for: %q", v.key)
			return
		}
	}

	keys := slices.MakeLookup(c.order)
	verify := func(c *cSource, method, name string) (err error) {
		switch name {
		case c.name, SourceIdKey:
			// c.name is an ad-hoc alias for source id column?
			return
		default:
			if _, present := keys[name]; !present {
				err = fmt.Errorf(".%s received unknown %q; known: %q", method, name, c.order)
				return
			}
		}
		return
	}

	for _, names := range c.unique {
		for _, name := range names {
			if err = verify(c, "AddUnique", name); err != nil {
				return
			}
		}
	}

	for _, names := range c.indexes {
		for _, name := range names {
			if err = verify(c, "AddIndex", name); err != nil {
				return
			}
		}
	}

	var t sqlbuilder.Table
	if t, err = c.getTable(); err != nil {
		return
	}
	configs, _ := c.getColumnConfigs()
	for _, cc := range configs {
		if key := t.C(cc.Name()); key != nil {
			c.keys[cc.Name()] = key
		}
	}

	err = c.idxs.finishAddSource(c)
	return
}

func (c *cSource) formal(more ...string) (name string) {
	//if c.parent != nil {
	//	return c.parent.formal(
	//		append(
	//			[]string{c.name},
	//			more...,
	//		)...,
	//	)
	//}
	return c.idxs.formal(c.name, more...)
}

func (c *cSource) getTable() (t sqlbuilder.Table, err error) {

	if c.table != nil {
		t = c.table
		return
	}

	var columns []sqlbuilder.ColumnConfig
	if columns, err = c.getColumnConfigs(); err != nil {
		return
	}

	t = c.idxs.b.NewTable(
		c.formal(),
		&sqlbuilder.TableOption{Unique: c.unique},
		columns...,
	)
	c.table = t

	return
}

func (c *cSource) getColumn(name string) (column sqlbuilder.Column, ok bool) {
	column, ok = c.keys[strcase.ToSnake(name)]
	return
}

func (c *cSource) getColumnConfig(name string) (config sqlbuilder.ColumnConfig, err error) {
	key := strcase.ToSnake(name)

	if cached, ok := c.column[key]; ok {
		config = cached
		return
	}

	switch key {
	case SourceIdKey:
		config = sqlbuilder.IntColumn(SourceIdKey, &sqlbuilder.ColumnOption{PrimaryKey: true, NotNull: true})
		c.column[key] = config
		return

	case c.value.key:
		if config, err = c.value.columnConfig(); err == nil {
			c.column[key] = config
		}
		return

	default:

		for _, v := range c.values {
			if v.key == key {
				config, err = v.columnConfig()
				c.column[key] = config
				return
			}
		}

	}

	err = fmt.Errorf("%w: %q", ErrColumnConfigNotFound, key)
	return
}

func (c *cSource) getColumnConfigs() (columns []sqlbuilder.ColumnConfig, err error) {
	for _, key := range append([]string{SourceIdKey}, c.order...) {
		if column, ee := c.getColumnConfig(key); ee == nil {
			columns = append(columns, column)
		} else {
			err = ee
			return
		}
	}
	return
}

func (c *cSource) IsPrimarySource() (yes bool) {
	yes = c.name == c.idxs.getPrimarySourceName()
	return
}

func (c *cSource) IsData() (yes bool) {
	yes = c.parent == nil && len(c.links) == 0
	return
}

func (c *cSource) IsLinked() (yes bool) {
	yes = c.parent != nil || len(c.links) > 0
	return
}

func (c *cSource) LinksTo() (names []string) {
	names = maps.SortedKeys(c.links)
	return
}

func (c *cSource) JoinLink(name string) (table sqlbuilder.Table, condition sqlbuilder.Condition, err error) {
	var ok bool
	var linkedKey string
	var linkedTable sqlbuilder.Table
	name = strcase.ToSnake(name)
	if linkedKey, ok = c.links[name]; !ok {
		err = fmt.Errorf("%q is not linked with %q", c.formal(), name)
		return
	} else if linkedTable, err = c.idxs.T(name); err != nil {
		err = fmt.Errorf("linked table not found: %q", name)
		return
	} else {
		table = linkedTable
		idKey := linkedTable.C(linkedKey) // foreign key
		thisColumn, _ := c.getColumn(name + "_" + linkedKey)
		condition = thisColumn.Eq(idKey)
	}
	return
}

func (c *cSource) makeJoinCond(otherName string) (other *cSource, cond sqlbuilder.Condition, err error) {
	var ok bool
	var table sqlbuilder.Table
	if table, err = c.getTable(); err == nil {
		if key, present := c.links[otherName]; present {
			if other, ok = c.idxs.getSource(otherName); ok {
				if otherColumn, ok := other.getColumn(key); ok {
					cond = otherColumn.Eq(table.C(otherName + "_" + key))
					return
				} else {
					err = fmt.Errorf("%s.makeJoinCond: %q.%q not found", c.name, otherName, key)
					return
				}
			} else {
				err = fmt.Errorf("%s.makeJoinCond: %q source not found", c.name, otherName)
			}
		} else {
			err = fmt.Errorf("%s.makeJoinCond: not linked to %q", c.name, otherName)
		}
	}
	return
}

func (c *cSource) makeJoinTable(otherName string) (joined sqlbuilder.Table, err error) {
	if other, cond, ee := c.makeJoinCond(otherName); ee != nil {
		err = ee
	} else {
		// self is top?
		joined = c.table.InnerJoin(other.table, cond)
	}
	return
}

func (c *cSource) JoinTable(with sqlbuilder.Table) (top sqlbuilder.Table, err error) {

	top = with // start with what was given

	thisTable, _ := c.getTable()

	var parentTableName string
	if c.parent != nil {

		// parent is present, join with this first
		var parentTable sqlbuilder.Table
		if parentTable, err = c.parent.getTable(); err != nil {
			return
		}
		parentTableName = parentTable.Name()
		top = top.InnerJoin(thisTable, parentTable.C(SourceIdKey).Eq(thisTable.C(c.parent.name+"_"+SourceIdKey)))

	} else if len(c.links) == 0 {
		// no parent and no links means no joining is necessary?
		// need to join with page because it's always present? not good
		// need to add this one to the FROM and not join
		return
	}

	for _, linkedTableName := range maps.SortedKeys(c.links) {
		if linkedTableName == parentTableName {
			continue
		}
		// join with each of the links, in natural order?
		if table, condition, ee := c.JoinLink(linkedTableName); ee != nil {
			err = ee
			return
		} else {
			top = top.InnerJoin(table, condition)
		}
	}

	return
}

func (c *cSource) MakeTable() (t sqlbuilder.Table, err error) {
	var top sqlbuilder.Table
	if c.parent != nil {
		if top, err = c.parent.getTable(); err != nil {
			return
		}
	} else if top, err = c.getTable(); err != nil {
		return
	}

	t, err = c.JoinTable(top)
	return
}
