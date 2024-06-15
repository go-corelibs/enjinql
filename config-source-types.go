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

type SourceConfigType uint8

const (
	UnknownSourceType SourceConfigType = iota
	// DataSourceType represents sources that have no parent and no linked values
	DataSourceType
	// LinkSourceType represents sources that have a parent or have linked values
	LinkSourceType
	// JoinSourceType represents sources that have a parent and have liked values
	JoinSourceType
)

func (t SourceConfigType) String() (name string) {
	switch t {
	case DataSourceType:
		return "data"
	case LinkSourceType:
		return "link"
	case JoinSourceType:
		return "join"
	default:
		return "unknown"
	}
}
