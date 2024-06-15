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

type OrderBy struct {
	Sources   *[]*SourceRef `parser:" 'ORDER' 'BY' (   @@ ( ',' @@ )*          " json:"key"`
	Random    *bool         `parser:"                | @( 'RANDOM' '(' ')' ) ) " json:"random,omitempty"`
	Direction *string       `parser:" @( 'ASC' | 'DSC' | 'DESC' )?             " json:"dir,omitempty"`

	Pos lexer.Position
}

func (o *OrderBy) IsDESC() bool {
	return o.Direction != nil && strings.ToUpper(*o.Direction) != "ASC"
}

func (o *OrderBy) make(state *cProcessor) (err error) {
	if err = o.validate(); err != nil {
		return
	}
	var columns []sqlbuilder.Column
	if o.Random != nil && *o.Random {
		columns = append(columns, sqlbuilder.Func("RANDOM"))
	} else {
		for _, srcRef := range *o.Sources {
			var column sqlbuilder.Column
			if column, err = srcRef.make(state); err != nil {
				return
			}
			columns = append(columns, column)
		}
	}
	state.build.OrderBy(o.IsDESC(), columns...)
	return
}

func (o *OrderBy) validate() (err error) {
	if o.Sources == nil {
		if o.Direction == nil {
			if o.Random == nil {
				return newSyntaxError(o.Pos, ErrInvalidSyntax, ErrNilStructure)
			}
		}
	}
	return
}

func (o *OrderBy) findSources() (names []*SrcKey) {
	if o.Sources != nil {
		for _, expr := range *o.Sources {
			names = append(names, expr.findSources()...)
		}
	}
	return
}

func (o *OrderBy) String() (out string) {
	if o.Random != nil && *o.Random {
		out += "RANDOM()"
	} else if o.Sources != nil {
		for _, expr := range *o.Sources {
			out += expr.String()
		}
	}
	if o.Direction != nil {
		out += " "
		if dir := strings.ToUpper(*o.Direction); dir == "DSC" {
			out += "DESC"
		} else {
			out += dir
		}
	}
	if out != "" {
		return "ORDER BY " + out
	}
	return
}
