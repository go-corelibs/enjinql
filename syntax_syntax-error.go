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
)

type SyntaxError struct {
	Pos      lexer.Position
	Parent   error
	Specific error
}

func newSyntaxError(pos lexer.Position, parent, err error) error {
	return &SyntaxError{
		Pos:      pos,
		Parent:   parent,
		Specific: err,
	}
}

func (e *SyntaxError) Error() string {
	return e.Err().Error()
}

func (e *SyntaxError) Err() error {
	if e.Parent != nil {
		return fmt.Errorf("%s %w: %w", e.Pos.String(), e.Parent, e.Specific)
	}
	return fmt.Errorf("%s %w", e.Pos.String(), e.Specific)
}
