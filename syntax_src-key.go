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

type SrcKey struct {
	Src   string
	Key   string
	Alias string
}

func newSrcKey(table, key, alias string) *SrcKey {
	return &SrcKey{
		Src:   table,
		Key:   key,
		Alias: alias,
	}
}

func (s *SrcKey) String() string {
	if s.Alias != "" {
		return s.Alias
	} else if s.Src == "" {
		return "." + s.Key
	}
	return s.Src + "." + s.Key
}
