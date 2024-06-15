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
)

type SourceKey struct {
	Source *string `parser:" ( @Ident (?= '.' ) )? " json:"source,omitempty"`
	Key    string  `parser:" '.' @Ident            " json:"key"`
	Alias  *string `parser:" ( 'AS' @Ident )?      " json:"alias,omitempty"`

	Pos lexer.Position
}

func (s *SourceKey) validate() (err error) {
	if s.Alias == nil {
		// not an alias, expecting at least key
		if s.Source == nil && s.Key == "" {
			return newSyntaxError(s.Pos, ErrInvalidSyntax, ErrNilStructure)
		} else if s.Key == "" {
			return newSyntaxError(s.Pos, ErrInvalidSyntax, ErrMissingSourceKey)
		}
	} else if *s.Alias == "" {
		return newSyntaxError(s.Pos, ErrInvalidSyntax, ErrNilStructure)
	}
	return
}

func (s *SourceKey) findSources() (names []*SrcKey) {
	var src, alias string
	if s.Source != nil {
		src = *s.Source
	}
	if s.Alias != nil {
		alias = *s.Alias
	}
	names = []*SrcKey{newSrcKey(src, s.Key, alias)}
	return
}

func (s *SourceKey) AsKey() (sk *SrcKey) {
	var src, alias string
	if s.Source != nil {
		src = *s.Source
	}
	if s.Alias != nil {
		alias = *s.Alias
	}
	return &SrcKey{
		Src:   src,
		Key:   s.Key,
		Alias: alias,
	}
}

func (s *SourceKey) String() (src string) {
	if s.Source != nil {
		src += *s.Source
	}
	src += "." + s.Key
	if s.Alias != nil {
		src += " AS " + *s.Alias
	}
	return src
}
