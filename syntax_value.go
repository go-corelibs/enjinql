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
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"

	clStrings "github.com/go-corelibs/strings"
)

type Value struct {
	Text        *string    `parser:"   @String                 " json:"text,omitempty"`
	Int         *int       `parser:" | @Int                    " json:"int,omitempty"`
	Float       *float64   `parser:" | @Float                  " json:"float,omitempty"`
	Bool        *Boolean   `parser:" | @( 'TRUE' | 'FALSE' )   " json:"bool,omitempty"`
	Null        *Null      `parser:" | @( 'NIL'  | 'NULL'  )   " json:"nil,omitempty"`
	SourceRef   *SourceRef `parser:" | @@                      " json:"source,omitempty"`
	Placeholder *string    `parser:" | @Placeholder            " json:"placeholder,omitempty"`

	Pos lexer.Position
}

// makeOther is for the right-hand side of a constraint expression
func (v *Value) makeOther(state *cProcessor) (other interface{}, err error) {

	switch {

	case v.SourceRef != nil:
		other, err = v.SourceRef.make(state)

	case v.Placeholder != nil:
		other = *v.Placeholder

	case v.Text != nil:
		text := *v.Text
		if size := len(text); size > 0 && text[0] == '\'' && text[size-1] == '\'' {
			text = clStrings.TrimQuotes(*v.Text)
			text = strings.ReplaceAll(text, `\"`, `"`)
			text = strings.ReplaceAll(text, `"`, `\"`)
			text = `"` + text + `"`
		}
		if other, err = strconv.Unquote(text); err != nil {
			err = newSyntaxError(v.Pos, ErrInvalidSyntax, fmt.Errorf("strconv.Unquote error: %w", err))
		}

	case v.Int != nil:
		other = *v.Int

	case v.Float != nil:
		other = *v.Float

	case v.Bool != nil:
		other = v.Bool.String()

	case v.Null != nil:
		other = v.Null.String()
	}

	return
}

func (v *Value) validate() (err error) {

	switch {
	case v.SourceRef != nil:
		return v.SourceRef.validate()
	case v.Placeholder != nil:
		return
	case v.Text != nil:
		return
	case v.Int != nil:
		return
	case v.Float != nil:
		return
	case v.Bool != nil:
		return
	case v.Null != nil:
		return
	}

	err = newSyntaxError(v.Pos, ErrInvalidSyntax, ErrNilStructure)
	return
}

func (v *Value) findSources() (sources []*SrcKey) {
	switch {
	case v.SourceRef != nil:
		return v.SourceRef.findSources()
	}
	return
}

func (v *Value) apply(argv ...interface{}) (err error) {
	if v.Placeholder != nil && *v.Placeholder != "" {
		var pos int
		number := *v.Placeholder
		number = number[1 : len(number)-1]
		if pos, err = strconv.Atoi(number); err == nil {
			pos -= 1
			if pos >= 0 && len(argv) > pos {
				switch t := argv[pos].(type) {
				case string:
					v.Text = &t
				case int:
					v.Int = &t
				case int8:
					i := int(t)
					v.Int = &i
				case int32:
					i := int(t)
					v.Int = &i
				case int64:
					i := int(t)
					v.Int = &i
				case float32:
					f := float64(t)
					v.Float = &f
				case float64:
					v.Float = &t
				case bool:
					b := Boolean(t)
					v.Bool = &b
				case nil:
					n := Null(true)
					v.Null = &n
				default:
					err = fmt.Errorf("%w: %q.(%T)", ErrSyntaxValueType, t, t)
					return
				}
				v.Placeholder = nil
			}
		}
	}
	return
}

func (v *Value) String() (out string) {

	switch {

	case v.SourceRef != nil:
		return v.SourceRef.String()

	case v.Placeholder != nil:
		return *v.Placeholder

	case v.Text != nil:
		return *v.Text

	case v.Int != nil:
		return strconv.Itoa(*v.Int)

	case v.Float != nil:
		return fmt.Sprintf("%v", *v.Float)

	case v.Bool != nil:
		return v.Bool.String()

	case v.Null != nil:
		return v.Null.String()

	}

	return
}
