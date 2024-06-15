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

	"github.com/alecthomas/participle/v2/lexer"
)

type Syntax struct {
	Lookup    bool         `parser:" ( ( @'LOOKUP'        " json:"lookup,omitempty"`
	Count     bool         `parser:"     @'COUNT'?        " json:"count,omitempty"`
	Distinct  bool         `parser:"     @'DISTINCT'?     " json:"distinct,omitempty"`
	Keys      []*SourceKey `parser:"     @@ ( ',' @@ )* ) " json:"keys,omitempty"`
	Query     bool         `parser:"   | @'QUERY' )       " json:"query,omitempty"`
	Within    *Expression  `parser:" ( 'WITHIN' @@ )?     " json:"within,omitempty"`
	OrderBy   *OrderBy     `parser:" ( @@ )?              " json:"orderBy,omitempty"`
	Offset    *int         `parser:" ( 'OFFSET' @Int )?   " json:"offset,omitempty"`
	Limit     *int         `parser:" ( 'LIMIT' @Int )?    " json:"limit,omitempty"`
	Semicolon bool         `parser:" ( @';' )?            " json:"semicolon,omitempty"`

	Pos lexer.Position
}

func (s *Syntax) init() (err error) {
	switch {
	case s.Lookup:
	case s.Count:
		if len(s.Keys) != 1 {
			err = fmt.Errorf("%w: COUNT requires exactly one context key", ErrInvalidSyntax)
			return
		}
	case s.Distinct:
		if len(s.Keys) != 1 {
			err = fmt.Errorf("%w: DISTINCT requires exactly one context key", ErrInvalidSyntax)
			return
		}
	case s.Query:
	default:
		if len(s.Keys) > 0 {
			s.Lookup = true
		} else {
			s.Query = true
		}
	}

	return
}

func (s *Syntax) String() string {
	var out string

	if s.Validate() == nil {

		switch {
		case s.Query:
			out += "QUERY"
		case s.Lookup:
			out += "LOOKUP"
		default:
			if len(s.Keys) > 0 {
				out += "LOOKUP"
			} else {
				out += "QUERY"
			}
		}

		if s.Lookup {

			if s.Count {
				out += " COUNT"
			}
			if s.Distinct {
				out += " DISTINCT"
			}

			if len(s.Keys) > 0 {
				for idx, sk := range s.Keys {
					if idx > 0 {
						out += ","
					}
					out += " " + sk.String()
				}
			}
		}

		if s.Within != nil {
			out += " WITHIN " + s.Within.String()
		}

		if s.OrderBy != nil {
			out += " " + s.OrderBy.String()
		}

		if s.Offset != nil {
			out += " " + strconv.Itoa(*s.Offset)
		}

		if s.Limit != nil {
			out += " " + strconv.Itoa(*s.Limit)
		}

		if s.Semicolon {
			out += ";"
		}
	}

	return out
}

func (s *Syntax) Validate() (err error) {
	numKeys := len(s.Keys)

	if s.Query {
		if numKeys > 0 {
			return newSyntaxError(s.Pos, ErrInvalidSyntax, ErrMismatchQuery)
		}
	} else if s.Lookup {
		if numKeys == 0 {
			return newSyntaxError(s.Pos, ErrInvalidSyntax, ErrMismatchLookup)
		} else if s.Count {
			if numKeys != 1 {
				err = fmt.Errorf("%w: COUNT requires exactly one source key", ErrInvalidSyntax)
				return
			}
		} else if s.Distinct {
			if numKeys != 1 {
				err = fmt.Errorf("%w: DISTINCT requires exactly one source key", ErrInvalidSyntax)
				return
			}
		}
	}

	for _, sk := range s.Keys {
		if err = sk.validate(); err != nil {
			return
		}
	}

	if s.Within != nil {
		if err = s.Within.validate(); err != nil {
			return
		}
	}

	if s.OrderBy != nil {
		if err = s.OrderBy.validate(); err != nil {
			return
		}
	}

	if s.Offset != nil {
		if *s.Offset < 0 {
			return newSyntaxError(s.Pos, ErrInvalidSyntax, ErrNegativeOffset)
		}
	}

	if s.Limit != nil {
		if *s.Limit < 0 {
			return newSyntaxError(s.Pos, ErrInvalidSyntax, ErrNegativeLimit)
		}
	}

	return
}

func (s *Syntax) findSources() (sources []*SrcKey) {
	for _, key := range s.Keys {
		sources = append(sources, key.findSources()...)
	}
	if s.Within != nil {
		sources = append(sources, s.Within.findSources()...)
	}
	if s.OrderBy != nil {
		sources = append(sources, s.OrderBy.findSources()...)
	}
	return
}

func (s *Syntax) findUpdatedSources() (sources []*SrcKey) {
	for _, key := range s.Keys {
		sources = append(sources, key.findSources()...)
	}
	if s.Within != nil {
		sources = append(sources, s.Within.findSources()...)
	}
	if s.OrderBy != nil {
		sources = append(sources, s.OrderBy.findSources()...)
	}
	return
}

func (s *Syntax) apply(argv ...interface{}) (err error) {
	if s.Within != nil {
		err = s.Within.apply(argv...)
	}
	return
}
