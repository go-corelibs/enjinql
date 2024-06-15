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
	"sync"

	mapset "github.com/deckarep/golang-set"
	"github.com/dominikbraun/graph"

	"github.com/go-corelibs/maps"
	"github.com/go-corelibs/slices"
)

/*

- a relationship is composed of two tables, joined on two columns
- need a type for defining a specific relationship


*/

// gSourceTableKey represents a specific <source>.<key> variable within a
// gSourceJoin statement, it is expected that the actual SQL generation phase
// will look up the correct source table
type gSourceTableKey struct {
	table string
	key   string
}

func newSourceTableKey(table, key string) gSourceTableKey {
	return gSourceTableKey{
		table: table,
		key:   key,
	}
}

func (g gSourceTableKey) String() string {
	return fmt.Sprintf("%s.%s", g.table, g.key)
}

// gSourceJoin represents an SQL INNER JOIN statement
//
//	INNER JOIN <table> ON <table>.<key> = <other>
type gSourceJoin struct {
	table string
	this  gSourceTableKey
	other gSourceTableKey
	note  string
}

func newSourceJoin(table, key string, other gSourceTableKey) *gSourceJoin {
	return &gSourceJoin{
		table: table,
		this:  gSourceTableKey{table: table, key: key},
		other: other,
	}
}

func (g *gSourceJoin) String() (output string) {
	output += fmt.Sprintf(`%v=%v`, g.other, g.this)
	return
}

type gSourcePlan struct {
	top   string
	joins []*gSourceJoin

	require []string
	topNote string
}

func newSourcePlan(top string) *gSourcePlan {
	return &gSourcePlan{top: top}
}

func (g *gSourcePlan) String() (out string) {
	out += "[" + g.top
	for _, join := range g.joins {
		out += ", "
		out += join.String()
	}
	out += "]"
	return
}

func (g *gSourcePlan) Verbose() (out string) {
	out += fmt.Sprintf("SRC\tquery sources\t%v\n", g.require)
	out += fmt.Sprintf("TOP\t%v\t%v\n", g.topNote, g.top)
	for idx, join := range g.joins {
		out += fmt.Sprintf("JOIN[%d]\tadd %v\t%v\n", idx+1, join.table, join.String())
	}
	return
}

func (g *gSourcePlan) Has(name string) (present bool) {
	if present = g.top == name; present {
		return
	}
	for _, join := range g.joins {
		if present = join.table == name; present {
			return
		}
	}
	return
}

func (g *gSourcePlan) add(join *gSourceJoin) {
	if !g.Has(join.table) {
		g.joins = append(g.joins, join)
	}
}

// gSourceNode represents a single source instance and tracks a mapping of
// other linked sources
type gSourceNode struct {
	// name of this source
	name string
	// parent is the specific join for the parent source, nil if there is no
	// parent
	parent *gSourceJoin
	// link is a mapping of other source names to specific table joins
	link map[string]*gSourceJoin
}

func newSourceNodeData(name string) *gSourceNode {
	return &gSourceNode{
		name: name,
		link: make(map[string]*gSourceJoin),
	}
}

func (g *gSourceNode) isData() bool {
	if g.parent != nil || len(g.link) > 0 {
		return false
	}
	return true
}

func (g *gSourceNode) join(other string) *gSourceJoin {
	if g.parent != nil && g.parent.other.table == other {
		return g.parent
	} else if j, ok := g.link[other]; ok {
		return j
	}
	return nil
}

func (g *gSourceNode) neighbors() (found []string) {
	if g.parent != nil {
		found = append(found, g.parent.other.table)
	}
	for _, link := range g.link {
		found = append(found, link.other.table)
	}
	return
}

type gSourceGraph struct {
	primary string
	nodes   []*gSourceNode
	lookup  map[string]*gSourceNode

	graph graph.Graph[string, *gSourceNode]

	m *sync.RWMutex
}

func gSourceHash(node *gSourceNode) (name string) {
	return node.name
}

func newSourceGraph() (g *gSourceGraph) {
	return &gSourceGraph{
		lookup: make(map[string]*gSourceNode),
		graph:  graph.New(gSourceHash),
		m:      &sync.RWMutex{},
	}
}

// Add distinct nodes only
func (g *gSourceGraph) Add(nodes ...*gSourceNode) (err error) {
	g.m.Lock()
	defer g.m.Unlock()

	for _, node := range nodes {
		if _, present := g.lookup[node.name]; !present {
			if len(g.nodes) == 0 {
				// the first added is considered the primary source
				g.primary = node.name
			}
			g.lookup[node.name] = node
			g.nodes = append(g.nodes, node)

			if err = g._addVertexUnsafe(node); err != nil {
				return
			}
		}
	}

	return
}

func (g *gSourceGraph) _addEdgeUnsafe(a, b string, join *gSourceJoin) (err error) {
	weight, data := graph.EdgeWeight(1), graph.EdgeData(join)
	if err = g.graph.AddEdge(a, b, weight, data); err != nil {
		return fmt.Errorf("error adding edge %q -> %q: %w", a, b, err)
	}
	return
}

func (g *gSourceGraph) _addVertexUnsafe(node *gSourceNode) (err error) {

	if err = g.graph.AddVertex(node); err != nil {
		return fmt.Errorf("error adding vertex %q: %w", node.name, err)
	}

	var parent string
	if node.parent != nil {
		// from parent to this
		parent = node.parent.other.table
		if err = g._addEdgeUnsafe(parent, node.name, node.parent); err != nil {
			return
		}
	}

	for _, link := range node.link {
		if err = g._addEdgeUnsafe(node.name, link.table, link); err != nil {
			return
		}
	}

	return
}

func (g *gSourceGraph) search(start, end string) (path []string, err error) {
	g.m.RLock()
	defer g.m.RUnlock()
	if path, err = graph.ShortestPath(g.graph, start, end); err != nil {
		err = fmt.Errorf("error searching from %q to %q: %w", start, end, err)
	}
	return
}

func (g *gSourceGraph) getNode(name string) (found *gSourceNode) {
	g.m.RLock()
	defer g.m.RUnlock()
	found, _ = g.lookup[name]
	return
}

func (g *gSourceGraph) validate() (err error) {
	g.m.RLock()
	defer g.m.RUnlock()

	pending := make(map[string]mapset.Set)

	for _, node := range g.nodes {
		deps := mapset.NewSet()
		if node.parent != nil {
			deps.Add(node.parent.other.table)
		}
		for name := range node.link {
			deps.Add(name)
		}
		pending[node.name] = deps
	}

	for len(pending) > 0 {
		// Get all nodes from the graph which have no dependencies
		empties := mapset.NewSet()
		for name, deps := range pending {
			if deps.Cardinality() == 0 {
				empties.Add(name)
			}
		}

		// If there aren't any ready nodes, then we have a circular dependency
		if empties.Cardinality() == 0 {
			return fmt.Errorf("circular dependency cycle: %v", maps.SortedKeys(pending))
		}

		// Remove the ready nodes and add them to the resolved graph
		for iter := range empties.Iter() {
			if name, ok := iter.(string); ok {
				delete(pending, name)
			}
		}

		// Also make sure to remove the ready nodes from the
		// remaining node dependencies as well
		for name, deps := range pending {
			pending[name] = deps.Difference(empties)
		}
	}

	return
}

func (g *gSourceGraph) plan(required ...string) (plan *gSourcePlan, err error) {
	var count int
	if count = len(required); count == 0 {
		// primary source is required if nothing else is
		if len(g.nodes) > 0 {
			// is it even possible to get here without a primary source?
			required = append(required, g.nodes[0].name)
		}
	}

	if err = g.validate(); err != nil {
		return
	} // no circular links detected

	pending := slices.NewStackUnique(required...)

	// confirm the sources requested are all present

	sources := make(map[string]*gSourceNode)
	for _, source := range pending.Slice() {
		if found := g.getNode(source); found == nil {
			err = fmt.Errorf("source not found: %q", source)
			return
		} else if _, present := sources[source]; !present {
			sources[source] = found
		}
	}

	g.m.RLock()
	defer g.m.RUnlock()

	// figure out the top source first
	var top, topNote string

	switch count {
	case 1:
		// only one source
		top, _ = pending.First()
		plan = newSourcePlan(top)
		plan.topNote = "only table"
		plan.require = required
		return
	}

	tops := slices.NewStackUnique[string]()
	parents := slices.NewStackUnique[string]()
	for _, source := range pending.Slice() {
		found := sources[source]
		if found.isData() {
			tops.Push(found.name)
		}
		if found.parent != nil {
			parents.Push(found.parent.other.table)
		}
	}

	switch tops.Len() {
	case 0:

		if parents.Len() == 1 {
			// they all share the same parent
			top, _ = parents.First()
			topNote = "first parent"
		} else {

			// prefer the primary source to be top
			if tops, pending, top = g.planPrimaryTopsUnsafe(tops, parents, pending); top == "" {
				// primary is not a parent nor present and no tops present
				// first source is top? not sure...
				top, _ = pending.Unshift()
				pending.Prune(top)
				topNote = "first required"
			} else {
				topNote = "primary source"
			}
		}

	default:
		// one or more tops present

		// prefer the primary source to be top
		if tops, pending, top = g.planPrimaryTopsUnsafe(tops, parents, pending); top == "" {
			// primary is not a parent nor present, first tops is top
			top, _ = tops.Unshift()
			pending.Prune(top)
			topNote = "first required"
		} else {
			topNote = "primary source"
		}

	}

	if top == "" {
		err = fmt.Errorf("failed to find a top table within: %v", required)
		return
	}

	plan = newSourcePlan(top)
	plan.topNote = topNote
	plan.require = required

	for pending.Len() > 0 {
		if source, ok := pending.Unshift(); ok {
			var path []string
			if path, err = g.search(top, source); err != nil {
				return
			}
			for idx, step := range path {
				if idx == 0 {
					// skip start, that's the top already
					continue
				}
				if edge, ee := g.graph.Edge(path[idx-1], step); ee == nil {
					if join, ok := edge.Properties.Data.(*gSourceJoin); ok {
						if present := plan.Has(join.table); !present {
							plan.add(join)
						}
					}
				}
			}
		}
	}

	if pending.Len() > 0 {
		err = fmt.Errorf("not enough constraints to resolve a plan including: %v, did plan: (%s)%v", pending.Slice(), top, plan.String())
		return
	}

	return
}

func (g *gSourceGraph) planPrimaryTopsUnsafe(tops, parents, sources *slices.StackUnique[string]) (*slices.StackUnique[string], *slices.StackUnique[string], string) {

	// is one of them the primary source? that's the top of the plan
	for _, source := range tops.Slice() {
		if source == g.primary {
			tops.Prune(source)
			sources.Prune(source)
			return tops, sources, source
		}
	}

	// none are the primary, are any of the parents the primary?
	for _, parent := range parents.Slice() {
		if parent == g.primary {
			tops.Prune(parent)
			sources.Prune(parent)
			return tops, sources, parent
		}
	}

	return tops, sources, ""
}
