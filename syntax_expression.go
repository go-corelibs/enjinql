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

type Expression struct {
	Constraint *Constraint `parser:"   @@  " json:"operation,omitempty"`
	Condition  *Condition  `parser:" | @@  " json:"condition,omitempty"`

	Pos lexer.Position
}

func (e *Expression) make(state *cProcessor) (cond sqlbuilder.Condition, err error) {

	if err = e.validate(); err != nil {
		return
	}

	if e.Condition != nil {
		// <expr> <AND|OR> <expr>
		if cond, err = e.Condition.make(state); err != nil {
			return
		}
	}

	if e.Constraint != nil {
		// <value> <operator> <value>
		if cond, err = e.Constraint.make(state); err != nil {
			return
		}
	}

	return
}

func (e *Expression) validate() (err error) {
	switch {
	case e.Condition != nil:
		return e.Condition.validate()
	case e.Constraint != nil:
		return e.Constraint.validate()
	}
	return newSyntaxError(e.Pos, ErrInvalidSyntax, ErrNilStructure)
}

func (e *Expression) findSources() (names []*SrcKey) {
	switch {
	case e.Condition != nil:
		names = e.Condition.findSources()
	case e.Constraint != nil:
		names = e.Constraint.findSources()
	}
	return
}

func (e *Expression) apply(argv ...interface{}) (err error) {
	switch {
	case e.Condition != nil:
		return e.Condition.apply(argv...)
	case e.Constraint != nil:
		return e.Constraint.apply(argv...)
	}
	return
}

func (e *Expression) String() (out string) {
	switch {
	case e.Condition != nil:
		return e.Condition.String()
	case e.Constraint != nil:
		return e.Constraint.String()
	}
	return
}
