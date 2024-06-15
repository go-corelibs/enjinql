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
	"errors"
)

var (
	ErrBuilderError    = errors.New("builder error")
	ErrInvalidSyntax   = errors.New("invalid syntax")
	ErrSyntaxValueType = errors.New("unsupported syntax value type")

	ErrNilStructure = errors.New("nil structure")

	ErrMismatchQuery  = errors.New("QUERY does not return keyed values; use LOOKUP for context specifics")
	ErrMismatchLookup = errors.New("LOOKUP does not return entire pages; use QUERY for complete pages")

	ErrNegativeOffset = errors.New("negative offset")
	ErrNegativeLimit  = errors.New("negative limit")

	ErrMissingSourceKey = errors.New("missing source key")
	ErrMissingOperator  = errors.New("missing operator")
	ErrMissingLeftSide  = errors.New("missing left-hand side expression")
	ErrMissingRightSide = errors.New("missing right-hand side expression")

	ErrInvalidConstraint = errors.New("invalid constraint")

	ErrInvalidInOp = errors.New("<SourceKey> [NOT] IN (<list>...)")

	ErrOpStringRequired = errors.New("operator requires a string argument")

	ErrTableNotFound  = errors.New("table not found")
	ErrColumnNotFound = errors.New("column not found")

	ErrInvalidJSON          = errors.New("invalid json data")
	ErrInvalidConfig        = errors.New("invalid config")
	ErrNoSources            = errors.New("at least one source is required")
	ErrNoSourceValues       = errors.New("at least one source value is required")
	ErrParentNotFound       = errors.New("parent not found")
	ErrNotSnakeCased        = errors.New("all names and keys must be snake_cased")
	ErrUnnamedSource        = errors.New("unnamed source")
	ErrEmptySourceValue     = errors.New("empty source value")
	ErrEmptySourceValueKey  = errors.New("source value key is empty")
	ErrSourceNotFound       = errors.New("source not found")
	ErrColumnConfigNotFound = errors.New("column config not found")
	ErrCreateIndexSQL       = errors.New("error building create index sql")
	ErrCreateIndex          = errors.New("error creating index sql")
	ErrCreateTableSQL       = errors.New("error building create table sql")
	ErrCreateTable          = errors.New("error creating table sql")

	ErrQueryRequiresStub = errors.New("eql query statements require a \"stub\" column")

	ErrDeleteRows    = errors.New("delete rows error")
	ErrInsertRow     = errors.New("insert row error")
	ErrTooManyValues = errors.New("too many values given")
	ErrNoValues      = errors.New("at least the first column value is required")
	ErrInvalidID     = errors.New("row identifiers must be greater than zero")

	ErrUnmarshalEnjinQL = errors.New("use enjinql.ParseConfig and enjinql.New to restore an EnjinQL instance")
)
