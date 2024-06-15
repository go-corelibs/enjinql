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

	"github.com/alecthomas/participle/v2/lexer"

	"github.com/go-corelibs/go-sqlbuilder"
)

type SourceRef struct {
	Source *string `parser:"   ( ( @Ident (?= '.' ) )?     " json:"source,omitempty"`
	Key    *string `parser:"     '.' @Ident            )   " json:"key,omitempty"`
	Alias  *string `parser:"   | @Ident                    " json:"alias,omitempty"`

	Pos lexer.Position
}

func (s *SourceRef) make(state *cProcessor) (c sqlbuilder.Column, err error) {
	if u, ok := state.updated[s.String()]; ok {
		c = u.c
	} else {
		err = fmt.Errorf("unknown source reference: %q", s.String())
	}
	return
}

func (s *SourceRef) validate() (err error) {
	if s.Alias == nil {
		// not an alias, expecting at least key
		if s.Source == nil && s.Key == nil {
			return newSyntaxError(s.Pos, ErrInvalidSyntax, ErrNilStructure)
		} else if s.Key == nil {
			return newSyntaxError(s.Pos, ErrInvalidSyntax, ErrMissingSourceKey)
		}
	} else if *s.Alias == "" {
		return newSyntaxError(s.Pos, ErrInvalidSyntax, ErrNilStructure)
	}
	return
}

func (s *SourceRef) findSources() (names []*SrcKey) {
	if s.Key == nil {
		// aliases reference other source instances
		// missing a key reference is an error
		return
	}
	var src, alias string
	if s.Source != nil {
		src = *s.Source
	}
	if s.Alias != nil {
		alias = *s.Alias
	}
	names = []*SrcKey{newSrcKey(src, *s.Key, alias)}
	return
}

func (s *SourceRef) String() string {
	switch {
	case s.Alias != nil:
		return *s.Alias
	case s.Source != nil && s.Key != nil:
		return fmt.Sprintf("%s.%s", *s.Source, *s.Key)
	case s.Source == nil && s.Key != nil:
		return "." + *s.Key
	}
	return ""
}
