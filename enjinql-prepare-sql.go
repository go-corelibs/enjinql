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
	"github.com/go-corelibs/go-sqlbuilder"
	"github.com/go-corelibs/values"
)

func (eql *enjinql) prepareSyntaxBuild(syntax *Syntax) (state *cProcessor, err error) {
	if err = syntax.Validate(); err != nil {
		return
	}

	if syntax.Query {
		// primary source "stub" is always the query context key
		// TODO: make the query stub context key configurable somehow
		if primarySource, ok := eql.sources.getPrimarySource(); ok {
			syntax.Keys = []*SourceKey{{
				Source: values.Ref(primarySource.name),
				Key:    PageStubKey,
				Alias:  nil,
				Pos:    syntax.Pos,
			}}
		}
	}

	state = &cProcessor{
		builder: eql.builder,
		syntax:  syntax,
		sources: eql.sources,
		tables:  make(map[string]sqlbuilder.Table),
		updated: make(map[string]*cProcessSrcKey),
	}

	//var order []string
	if _ /*order*/, state.updated, err = state.findUpdatedSrcKeyRefs(); err != nil {
		return
	}

	return
}

func (eql *enjinql) preparePlan(syntax *Syntax) (plan *gSourcePlan, err error) {
	var state *cProcessor
	if state, err = eql.prepareSyntaxBuild(syntax); err != nil {
		return
	} else if plan, err = state.preparePlan(); err != nil {
		return
	}
	return
}

func (eql *enjinql) prepareSQL(syntax *Syntax) (sql string, argv []interface{}, err error) {
	var state *cProcessor
	if state, err = eql.prepareSyntaxBuild(syntax); err != nil {
		return
	}

	primarySourceName := eql.sources.getPrimarySourceName()

	var top sqlbuilder.Table
	if top, err = state.prepareBuild(); err != nil {
		// TODO: is testing prepareBuild here necessary?
		return
	}

	state.build = eql.builder.Select(top)

	getColumn := func(sk *SourceKey) (column sqlbuilder.Column, alias string, ok bool) {
		var bsk *cProcessSrcKey
		if ok = sk.Alias != nil; ok {
			if bsk, ok = state.updated[*sk.Alias]; ok {
				column = bsk.c
				alias = *sk.Alias
				return
			}
		} else if bsk, ok = state.updated[sk.String()]; ok {
			column = bsk.c
			return
		}
		return
	}

	if state.syntax.Lookup {
		// lookup <columns> within <expression> order...
		// select <columns> from <table> <joins> where <expression> order by <expression> offset <int> limit <int>

		var columns []sqlbuilder.Column

		// syntax.Validate ensures specifically one column present for COUNT and DISTINCT statements
		switch {
		case state.syntax.Count && state.syntax.Distinct:
			if c, alias, ok := getColumn(state.syntax.Keys[0]); ok {
				fn := sqlbuilder.Func(
					"COUNT",
					sqlbuilder.Func("DISTINCT", c),
				)
				if alias != "" {
					columns = []sqlbuilder.Column{fn.As(alias)}
				} else {
					columns = []sqlbuilder.Column{fn}
				}
			}
		case state.syntax.Count:
			if c, alias, ok := getColumn(state.syntax.Keys[0]); ok {
				fn := sqlbuilder.Func("COUNT", c)
				if alias != "" {
					columns = []sqlbuilder.Column{fn.As(alias)}
				} else {
					columns = []sqlbuilder.Column{fn}
				}
			}
		case state.syntax.Distinct:
			if c, alias, ok := getColumn(state.syntax.Keys[0]); ok {
				fn := sqlbuilder.Func("DISTINCT", c)
				if alias != "" {
					columns = []sqlbuilder.Column{fn.As(alias)}
				} else {
					columns = []sqlbuilder.Column{fn}
				}
			}
		default:
			for _, sk := range state.syntax.Keys {
				if column, alias, ok := getColumn(sk); ok {
					if alias != "" {
						columns = append(columns, column.As(alias))
						continue
					}
					columns = append(columns, column)
				}
			}
		}

		state.build.Columns(columns...)

	} else if state.syntax.Query {
		// query within <expression> order...
		// select <page>.stub from <page> <joins> where <expression> order by <expression> offset <int> limit <int>

		var ok bool
		var source *cSource
		var t sqlbuilder.Table
		if source, ok = eql.sources.getSource(primarySourceName); ok {
			if t, err = source.getTable(); err != nil {
				return
			} else if stub := t.C(PageStubKey); stub != nil {
				// TODO: need a means of specifying the "stub" column in a source config so that non PageSource setups can work
				state.build.Columns(stub)
			} else {
				err = ErrQueryRequiresStub
				return
			}
		} else {
			err = ErrSourceNotFound
			return
		}

	} // state.prepareBuild already validated the !Lookup && !Query case

	if state.syntax.Within != nil {

		var cond sqlbuilder.Condition
		if cond, err = state.syntax.Within.make(state); err != nil {
			return
		}
		state.build.Where(cond)

	}

	if state.syntax.OrderBy != nil {
		if err = state.syntax.OrderBy.make(state); err != nil {
			return
		}
	}

	if state.syntax.Offset != nil {
		state.build.Offset(*state.syntax.Offset)
	}

	if state.syntax.Limit != nil {
		state.build.Limit(*state.syntax.Limit)
	}

	return state.build.ToSql()
}
