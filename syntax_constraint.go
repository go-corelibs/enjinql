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
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/go-corelibs/go-sqlbuilder"
)

// Constraint is the comparing of two values
type Constraint struct {
	Left   *SourceRef `parser:" @@                               " json:"left"`
	Op     *Operator  `parser:" (   ( @@                         " json:"op,omitempty"`
	Right  *Value     `parser:"       @@ )                       " json:"right,omitempty"`
	Not    bool       `parser:"   | ( @'NOT'?                    " json:"not,omitempty"`
	In     bool       `parser:"       @'IN'                      " json:"in,omitempty"`
	Values []*Value   `parser:"       '(' @@ ( ',' @@ )* ')' ) ) " json:"values,omitempty"`

	Pos lexer.Position
}

func (c *Constraint) make(state *cProcessor) (cond sqlbuilder.Condition, err error) {
	var src sqlbuilder.Column
	var other interface{}

	if c.Left == nil || (c.Op == nil && !c.In) {
		// left is nil, or op is nil and not IN either
		err = newSyntaxError(c.Pos, ErrInvalidSyntax, ErrInvalidConstraint)
		return
	}

	if src, err = c.Left.make(state); err != nil {
		return
	}

	if c.In {
		// src.Ref NOT? IN ( <values> )
		var values []interface{}
		for _, value := range c.Values {
			if other, err = value.makeOther(state); err != nil {
				err = newSyntaxError(c.Pos, ErrInvalidSyntax, err)
				return
			} else {
				values = append(values, other)
			}
		}
		if c.Not {
			cond = src.NotIn(values...)
			return
		}
		cond = src.In(values...)
		return

	}

	// src.Ref <op> <value>
	if other, err = c.Right.makeOther(state); err != nil {
		return
	}

	cond, err = c.Op.make(src, other)
	return
}

func (c *Constraint) apply(argv ...interface{}) (err error) {
	// c.Left is a source ref, no placeholder
	if c.Right != nil {
		err = c.Right.apply(argv...)
	}
	return
}

func (c *Constraint) String() (out string) {
	if c.validate() == nil {
		out += c.Left.String()

		if c.In {
			if c.Not {
				out += " NOT"
			}
			out += " IN ("
			for idx, value := range c.Values {
				if idx > 0 {
					out += ", "
				}
				out += value.String()
			}
			out += ")"
			return
		}

		out += " "
		out += c.Op.String()
		out += " "
		out += c.Right.String()
	}
	return
}

func (c *Constraint) validate() (err error) {

	// double-check the left-hand side
	if c.Left == nil {
		return newSyntaxError(c.Pos, ErrInvalidSyntax, ErrMissingLeftSide)
	} else if err = c.Left.validate(); err != nil {
		return
	}

	if c.In {

		if len(c.Values) == 0 {
			return newSyntaxError(c.Pos, ErrInvalidSyntax, ErrInvalidInOp)
		}

		for _, value := range c.Values {
			if err = value.validate(); err != nil {
				return
			}
		}

		return
	}

	if c.Op == nil {
		return newSyntaxError(c.Pos, ErrInvalidSyntax, ErrMissingOperator)
	} else if c.Right == nil {
		return newSyntaxError(c.Pos, ErrInvalidSyntax, ErrMissingRightSide)
	}

	return
}

func (c *Constraint) findSources() (names []*SrcKey) {
	if c.Left != nil {
		names = append(names, c.Left.findSources()...)
	}
	if c.In {
		for _, value := range c.Values {
			names = append(names, value.findSources()...)
		}
		return
	}

	if c.Right != nil {
		names = append(names, c.Right.findSources()...)
	}
	return
}
