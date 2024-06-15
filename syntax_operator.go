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

// Operator represents a comparison operation
//
//	| Key |  Op  | Description              |
//	+-----+------+--------------------------+
//	| EQ  |  ==  | equal to                 |
//	| NE  |  !=  | not equal to             |
//	| GE  |  >=  | greater than or equal to |
//	| LE  |  <=  | less than or equal to    |
//	| GT  |  >   | greater than             |
//	| LT  |  <   | less than                |
//	| LK  | LIKE | like                     |
//	| SW  |  ^=  | starts with              |
//	| EW  |  $=  | ends with                |
//	| CS  |  *=  | contains one of string   |
//	| CF  |  ~=  | contains any of fields   |
//
// For LK, SW, EW, CS and CF, there is a NOT modifier:
//
//	| Key |  Op  | Description              |
//	+-----+------+--------------------------+
//	| Not |  NOT | long-form negate         |
//	| Nt  |  !   | short-form negate        |
//
// Example NOT modifier usage:
//
//	| Example | Description         |
//	+---------+---------------------+
//	| NOT ^=  | does not start with |
//	|   !$=   | does not end with   |
type Operator struct {
	EQ  bool `parser:" (   @'=='               " json:"eq,omitempty"`
	NE  bool `parser:"   | @( '!=' | '<>' )    " json:"ne,omitempty"`
	GE  bool `parser:"   | @'>='               " json:"ge,omitempty"`
	LE  bool `parser:"   | @'<='               " json:"le,omitempty"`
	GT  bool `parser:"   | @'>'                " json:"gt,omitempty"`
	LT  bool `parser:"   | @'<'                " json:"lt,omitempty"`
	Not bool `parser:" ) | ( (   @( 'NOT' )    " json:"not,omitempty"`
	Nt  bool `parser:"         | @( '!' )   )? " json:"nt,omitempty"`
	LK  bool `parser:"     (   @'LIKE'         " json:"lk,omitempty"`
	SW  bool `parser:"       | @'^='           " json:"sw,omitempty"`
	EW  bool `parser:"       | @'$='           " json:"ew,omitempty"`
	CS  bool `parser:"       | @'*='           " json:"cs,omitempty"`
	CF  bool `parser:"       | @'~='       ) ) " json:"cf,omitempty"`

	Pos lexer.Position
}

func (o Operator) String() string {
	var out string
	if o.Not {
		out += "NOT "
	} else if o.Nt {
		out += "!"
	}

	switch {
	case o.EQ:
		return out + "=="
	case o.NE:
		return out + "!="
	case o.LE:
		return out + "<="
	case o.GE:
		return out + ">="
	case o.LT:
		return out + "<"
	case o.GT:
		return out + ">"

	case o.LK:
		return out + "LIKE"
	case o.SW:
		return out + "^="
	case o.EW:
		return out + "$="
	case o.CS:
		return out + "*="
	case o.CF:
		return out + "~="
	}

	return ""
}

func (o Operator) validate() (err error) {
	if o.String() == "" {
		return newSyntaxError(o.Pos, ErrInvalidSyntax, ErrNilStructure)
	}
	return
}

func (o Operator) make(c sqlbuilder.Column, right interface{}) (cond sqlbuilder.Condition, err error) {
	if err = o.validate(); err == nil {
		switch {
		case o.EQ:
			cond = c.Eq(right)
		case o.NE:
			cond = c.NotEq(right)
		case o.LE:
			cond = c.LtEq(right)
		case o.GE:
			cond = c.GtEq(right)
		case o.LT:
			cond = c.Lt(right)
		case o.GT:
			cond = c.Gt(right)

		case o.CS: // *= contains string
			return o.makeCS(c, right)
		case o.CF: // ~= contains field (at least one)
			return o.makeCF(c, right)
		case o.SW: // ^= starts with
			return o.makeSW(c, right)
		case o.EW: // $= ends with
			return o.makeEW(c, right)
		case o.LK: // is like
			return o.makeLK(c, right)

		}
	}
	return
}

func (o Operator) makeCS(c sqlbuilder.Column, right interface{}) (cond sqlbuilder.Condition, err error) {
	if v, ok := right.(string); ok {
		if o.Not || o.Nt {
			cond = c.NotLike("%" + v + "%")
		} else {
			cond = c.Like("%" + v + "%")
		}
		return
	}
	err = newSyntaxError(o.Pos, ErrInvalidSyntax, ErrOpStringRequired)
	return
}

func (o Operator) makeCF(c sqlbuilder.Column, right interface{}) (cond sqlbuilder.Condition, err error) {
	if v, ok := right.(string); ok {
		fields := strings.Fields(v)
		var conditions []sqlbuilder.Condition
		for _, field := range fields {
			if o.Not || o.Nt {
				conditions = append(conditions, c.NotLike("%"+field+"%"))
			} else {
				conditions = append(conditions, c.Like("%"+field+"%"))
			}
		}
		if len(conditions) == 0 {
			err = newSyntaxError(o.Pos, ErrInvalidSyntax, ErrOpStringRequired)
			return
		}
		cond = sqlbuilder.Or(conditions...)
		return
	}
	err = newSyntaxError(o.Pos, ErrInvalidSyntax, ErrOpStringRequired)
	return
}

func (o Operator) makeSW(c sqlbuilder.Column, right interface{}) (cond sqlbuilder.Condition, err error) {
	if v, ok := right.(string); ok {
		if o.Not || o.Nt {
			cond = c.NotLike(v + "%")
		} else {
			cond = c.Like(v + "%")
		}
		return
	}
	err = newSyntaxError(o.Pos, ErrInvalidSyntax, ErrOpStringRequired)
	return
}

func (o Operator) makeEW(c sqlbuilder.Column, right interface{}) (cond sqlbuilder.Condition, err error) {
	if v, ok := right.(string); ok {
		if o.Not || o.Nt {
			cond = c.NotLike("%" + v)
		} else {
			cond = c.Like("%" + v)
		}
		return
	}
	err = newSyntaxError(o.Pos, ErrInvalidSyntax, ErrOpStringRequired)
	return
}

func (o Operator) makeLK(c sqlbuilder.Column, right interface{}) (cond sqlbuilder.Condition, err error) {
	if v, ok := right.(string); ok {
		if o.Not || o.Nt {
			cond = c.NotLike(v)
		} else {
			cond = c.Like(v)
		}
		return
	}
	err = newSyntaxError(o.Pos, ErrInvalidSyntax, ErrOpStringRequired)
	return
}
