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
	"strings"
	"sync"

	"github.com/iancoleman/strcase"

	"github.com/go-corelibs/go-sqlbuilder"
	"github.com/go-corelibs/maps"
	"github.com/go-corelibs/slices"
)

type cSources struct {
	prefix string              // snake_cased table prefix
	order  []string            // order the sources were added
	lookup map[string]*cSource // formal_name -> source
	b      sqlbuilder.Buildable

	graph *gSourceGraph

	sync.RWMutex
}

func newSources(prefix string, b sqlbuilder.Buildable) *cSources {
	return &cSources{
		prefix: strcase.ToSnake(prefix),
		lookup: make(map[string]*cSource),
		graph:  newSourceGraph(),
		b:      b,
	}
}

// alias returns the given name with the enjinql prefix removed
func (c *cSources) alias(name string) string {
	name = strcase.ToSnake(name)
	if c.prefix == "" {
		return name
	}
	return strings.TrimPrefix(c.formal(name), c.prefix+"_")
}

// formal returns the full name, including the enjinql prefix and any more
// given, in snake_cased format
func (c *cSources) formal(name string, more ...string) (full string) {

	name = strcase.ToSnake(name)
	for idx := range more {
		more[idx] = strcase.ToSnake(more[idx])
	}

	if full = name; len(more) > 0 {
		full += "_" + strings.Join(more, "_")
	}

	if c.prefix == "" {
		// no prefix
		return full
	} else if full != c.prefix && strings.HasPrefix(full, c.prefix+"_") {
		// has exact prefix already
		return full
	}

	parts := strings.Split(c.prefix, "_")
	for i := len(parts) - 1; i >= 0; i-- {
		tmp := strings.Join(parts[i:], "_") + "_"
		if strings.HasPrefix(full, tmp) {
			full = strings.TrimPrefix(full, tmp)
			return c.prefix + "_" + full
		}
	}

	return c.prefix + "_" + full
}

func (c *cSources) exists(name string) (present bool) {
	c.RLock()
	defer c.RUnlock()
	//_, present = c.lookup[c.formal(name)]
	_, present = c.lookup[name]
	return
}

func (c *cSources) addSource(sc *SourceConfig) (err error) {
	if c.exists(sc.Name) {
		err = fmt.Errorf("%q source exists already", sc.Name)
		return
	}

	source := &cSource{
		name:    sc.Name,
		node:    &gSourceNode{name: sc.Name, link: make(map[string]*gSourceJoin)},
		idxs:    c,
		keys:    make(map[string]sqlbuilder.Column),
		order:   make([]string, 0),
		links:   make(map[string]string),
		values:  make([]cSourceValue, 0),
		unique:  make([][]string, 0),
		indexes: make([][]string, 0),
		column:  make(map[string]sqlbuilder.ColumnConfig),
	}

	if sc.Name == "" {
		err = fmt.Errorf("%w: %w", ErrInvalidConfig, ErrUnnamedSource)
		return
	}

	if sc.Parent != nil {
		if parent, present := c.lookup[*sc.Parent]; !present {
			err = fmt.Errorf("parent soruce not found: %q", *sc.Parent)
			return
		} else {
			source.parent = parent
			source.node.parent = newSourceJoin(
				sc.Name, *sc.Parent+"_"+SourceIdKey,
				newSourceTableKey(parent.name, SourceIdKey),
			)
		}
	}

	for idx, value := range sc.Values {
		switch {
		case value.Int != nil:
			source.values = append(source.values, cSourceValue{
				ivt: gIntValue,
				key: value.Int.Key,
				opt: &sqlbuilder.ColumnOption{},
			})
			source.order = append(source.order, value.Int.Key)
		case value.Bool != nil:
			source.values = append(source.values, cSourceValue{
				ivt: gBoolValue,
				key: value.Bool.Key,
				opt: &sqlbuilder.ColumnOption{},
			})
			source.order = append(source.order, value.Bool.Key)
		case value.Time != nil:
			source.values = append(source.values, cSourceValue{
				ivt: gTimeValue,
				key: value.Time.Key,
				opt: &sqlbuilder.ColumnOption{},
			})
			source.order = append(source.order, value.Time.Key)
		case value.Float != nil:
			source.values = append(source.values, cSourceValue{
				ivt: gFloatValue,
				key: value.Float.Key,
				opt: &sqlbuilder.ColumnOption{},
			})
			source.order = append(source.order, value.Float.Key)
		case value.String != nil:
			opt := &sqlbuilder.ColumnOption{}
			if value.String.Size > 0 {
				opt.Size = value.String.Size
			}
			source.values = append(source.values, cSourceValue{
				ivt: gStringValue,
				key: value.String.Key,
				opt: opt,
			})
			source.order = append(source.order, value.String.Key)
		case value.Linked != nil:
			linkedKey := value.Linked.Source + "_" + value.Linked.Key
			source.values = append(source.values, cSourceValue{
				ivt: gLinkValue,
				key: linkedKey,
				opt: &sqlbuilder.ColumnOption{NotNull: true},
			})
			source.order = append(source.order, linkedKey)
			if c.getPrimarySourceName() != value.Linked.Source {
				source.node.link[value.Linked.Source] = newSourceJoin(
					value.Linked.Source, value.Linked.Key,
					newSourceTableKey(sc.Name, value.Linked.Source+"_"+SourceIdKey),
				)
			}
		default:
			err = fmt.Errorf("%w: %w (%q value #%d)", ErrInvalidConfig, ErrEmptySourceValue, sc.Name, idx)
			return
		}
	}

	if len(source.values) == 0 && source.parent == nil {
		err = fmt.Errorf("%w: %w", ErrInvalidConfig, ErrNoSourceValues)
		return
	}

	if source.parent == nil {
		// the first value is the primary value
		source.values, source.value = slices.Unshift(source.values)
		//source.order, _ = slices.Unshift(source.order)
	} else {
		// parent link is the primary value
		linkedKey := source.parent.name + "_" + SourceIdKey
		source.value = cSourceValue{
			ivt: gLinkValue,
			key: linkedKey,
			opt: &sqlbuilder.ColumnOption{NotNull: true},
		}
		source.order = slices.Shift(source.order, linkedKey)
		if c.getPrimarySourceName() != source.parent.name {
			source.links[source.parent.name] = SourceIdKey
		}
	}

	source.unique = sc.Unique
	source.indexes = sc.Index

	if err = c.graph.Add(source.node); err != nil {
		return fmt.Errorf("error adding node to graph: %q - %w", source.node.name, err)
	}

	err = source.init()
	return
}

func (c *cSources) finishAddSource(source *cSource) (err error) {
	name := source.name // source.formal()
	c.RLock()
	if _, present := c.lookup[name]; present {
		err = fmt.Errorf("%q source exists already", name)
		c.RUnlock()
		return
	}
	c.RUnlock()
	c.Lock()
	defer c.Unlock()

	// add to the primary lookup
	c.lookup[name] = source
	// track the order of sources added
	c.order = append(c.order, name)
	return
}

func (c *cSources) getPrimarySource() (source *cSource, ok bool) {
	c.RLock()
	defer c.RUnlock()
	if ok = len(c.order) > 0; ok {
		//source, ok = c.lookup[c.formal(c.order[0])]
		source, ok = c.lookup[c.order[0]]
	}
	return
}

func (c *cSources) getPrimarySourceName() (name string) {
	if source, ok := c.getPrimarySource(); ok {
		name = source.name
	}
	return
}

func (c *cSources) getPrimarySourceFormal() (formal string) {
	if source, ok := c.getPrimarySource(); ok {
		formal = source.formal()
	}
	return
}

func (c *cSources) getSource(name string) (source *cSource, ok bool) {
	c.RLock()
	defer c.RUnlock()
	//source, ok = c.lookup[c.formal(name)]
	source, ok = c.lookup[name]
	return
}

func (c *cSources) T(name string) (table sqlbuilder.Table, err error) {
	if s, ok := c.getSource(name); ok {
		table, err = s.getTable()
	} else {
		err = fmt.Errorf("source not found: %q", name)
	}
	return
}

func (c *cSources) listSources() (names []string) {
	c.RLock()
	defer c.RUnlock()
	names = maps.SortedKeys(c.lookup)
	return
}
