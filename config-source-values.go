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

// ConfigSourceValues is a slice type for providing Config builder methods
type ConfigSourceValues []*SourceConfigValue

func (v ConfigSourceValues) update(c *Config) {
	for _, value := range v {
		value.update(c)
	}
}

func (v ConfigSourceValues) Clone() (cloned ConfigSourceValues) {
	for _, value := range v {
		cloned = append(cloned, value.Clone())
	}
	return
}

func (v ConfigSourceValues) Names() (names []string) {
	for _, value := range v {
		names = append(names, value.Name())
	}
	return
}

func (v ConfigSourceValues) HasLinks() (linked bool) {
	for _, value := range v {
		if linked = value.Linked != nil; linked {
			return
		}
	}
	return
}
