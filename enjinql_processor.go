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

type cProcessSrcKey struct {
	name string

	k bool
	s *cSource
	t sqlbuilder.Table
	c sqlbuilder.Column
	o *SrcKey
	u *SrcKey
}

type cProcessor struct {
	builder sqlbuilder.Buildable
	build   sqlbuilder.SelectBuilder
	syntax  *Syntax
	tables  map[string]sqlbuilder.Table
	sources *cSources
	order   []string
	updated map[string]*cProcessSrcKey
}

func (p *cProcessor) findUpdatedSrcKeyRefs() (order []string, updated map[string]*cProcessSrcKey, err error) {

	ctxKeys := make(map[string]struct{})
	aliased := make(map[string]*SrcKey)

	stack := slices.NewStackUnique[string]()
	primarySourceName := p.sources.getPrimarySourceName()

	for _, sk := range p.syntax.Keys {
		if sk.Source == nil || *sk.Source == "" {
			ctxKeys[primarySourceName] = struct{}{}
		} else if src, ok := p.sources.getSource(*sk.Source); ok {
			ctxKeys[src.formal()] = struct{}{}
		}
		if sk.Alias != nil {
			aliased[*sk.Alias] = sk.AsKey()
		}
	}

	updated = make(map[string]*cProcessSrcKey)
	for _, found := range p.syntax.findSources() {
		if found.Alias != "" {
			if v, ok := aliased[found.Alias]; ok {
				found = v
			} else {
				err = fmt.Errorf("unknown source key alias: %q", found.Alias)
				return
			}
		}
		update := &SrcKey{
			Src: strcase.ToSnake(found.Src),
			Key: strcase.ToSnake(found.Key),
		}
		if found.Src == "" {
			update.Src = primarySourceName
		}

		if t, ee := p.sources.T(update.Src); ee != nil {
			err = fmt.Errorf("%w: %q", ErrTableNotFound, found.Src)
			return
		} else if column := t.C(strcase.ToSnake(update.Key)); sqlbuilder.IsColumnError(column) {
			err = fmt.Errorf("%w: %q.%q", ErrColumnNotFound, t.Name(), found.Key)
			return
		} else {
			if source, ok := p.sources.getSource(update.Src); !ok {
				err = fmt.Errorf("unknown source name: %q", found.Src)
				return
			} else if columnConfig, eee := source.getColumnConfig(update.Key); eee != nil {
				err = eee
				return
			} else {
				update.Src = source.formal()
				update.Key = columnConfig.Name()
				formal := found.String()
				_, isKey := ctxKeys[update.Src]
				stack.Push(formal)
				updated[formal] = &cProcessSrcKey{name: source.name, k: isKey, s: source, t: t, c: column, o: found, u: update}
				if found.Alias != "" {
					updated[found.Alias] = updated[formal]
				}
			}
		}
	}

	order = stack.Slice()
	return

}

func (p *cProcessor) getRequiredSources() (required []string, err error) {

	unique := make(map[string]struct{})

	for _, formal := range maps.SortedKeys(p.updated) {
		bsk := p.updated[formal]
		if _, present := unique[bsk.u.Src]; present {
			continue
		}
		unique[bsk.u.Src] = struct{}{}
		required = append(required, p.sources.alias(bsk.u.Src))
	}

	if len(required) == 0 {
		// there are no sources? how is this case even possible?
		err = fmt.Errorf("no sources required, strange")
	}
	return
}

func (p *cProcessor) preparePlan() (planned *gSourcePlan, err error) {
	var required []string
	if required, err = p.getRequiredSources(); err != nil {
		return
	} else if planned, err = p.sources.graph.plan(required...); err != nil {
		return
	}
	return
}

func (p *cProcessor) prepareBuild() (top sqlbuilder.Table, err error) {

	var planned *gSourcePlan
	if planned, err = p.preparePlan(); err != nil {
		return
	} else if source, ok := p.sources.getSource(planned.top); ok {
		if top, err = source.getTable(); err != nil {
			return
		}
	}

	for _, join := range planned.joins {
		if source, ok := p.sources.getSource(join.table); ok {
			var thisTable, otherTable sqlbuilder.Table
			if thisTable, err = source.getTable(); err != nil {
				return
			} else if otherSource, ok := p.sources.getSource(join.other.table); ok {
				if otherTable, err = otherSource.getTable(); err != nil {
					return
				}
			}
			if thisColumn := thisTable.C(join.this.key); thisColumn != nil {
				if otherColumn := otherTable.C(join.other.key); otherColumn != nil {
					top = top.InnerJoin(thisTable, otherColumn.Eq(thisColumn))
				}
			}

		}
	}

	return
}
