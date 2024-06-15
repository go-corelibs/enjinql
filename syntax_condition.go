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
	"strings"

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/go-corelibs/go-sqlbuilder"
)

// Condition is the AND/OR combining of two expressions
type Condition struct {
	Left  *Expression `parser:" '(' @@ ')'        " json:"left"`
	Type  string      `parser:" @( 'AND' | 'OR' ) " json:"type"`
	Right *Expression `parser:" '(' @@ ')'        " json:"right"`

	Pos lexer.Position
}

func (c *Condition) make(state *cProcessor) (cond sqlbuilder.Condition, err error) {
	var lc, rc sqlbuilder.Condition
	if lc, err = c.Left.make(state); err != nil {
		return
	} else if rc, err = c.Right.make(state); err != nil {
		return
	}
	switch strings.ToUpper(c.Type) {
	case "AND":
		cond = sqlbuilder.And(lc, rc)
	case "OR":
		cond = sqlbuilder.Or(lc, rc)
	default:
		// should never happen?
		err = newSyntaxError(c.Pos, ErrInvalidSyntax, ErrBuilderError)
	}
	return
}

func (c *Condition) validate() (err error) {
	if c.Left == nil && c.Right == nil {
		return newSyntaxError(c.Pos, ErrInvalidSyntax, ErrNilStructure)
	} else if c.Left == nil {
		return newSyntaxError(c.Pos, ErrInvalidSyntax, ErrMissingLeftSide)
	} else if err = c.Left.validate(); err != nil {
		return
	} else if c.Right == nil {
		return newSyntaxError(c.Pos, ErrInvalidSyntax, ErrMissingRightSide)
	} else if err = c.Right.validate(); err != nil {
		return
	}
	return
}

func (c *Condition) findSources() (names []*SrcKey) {
	if c.Left != nil {
		names = append(names, c.Left.findSources()...)
	}
	if c.Right != nil {
		names = append(names, c.Right.findSources()...)
	}
	return
}

func (c *Condition) apply(argv ...interface{}) (err error) {
	if c.Left != nil {
		if err = c.Left.apply(argv...); err != nil {
			return
		}
	}
	if c.Right != nil {
		err = c.Right.apply(argv...)
	}
	return
}

func (c *Condition) String() (out string) {
	if c.validate() == nil {
		out += "(" + c.Left.String() + ")"
		out += " "
		out += strings.ToUpper(c.Type)
		out += " "
		out += "(" + c.Right.String() + ")"
	}
	return
}
