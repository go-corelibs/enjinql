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

// MakeSourceConfig is a convenience wrapper for constructing SourceConfig
// instances. Will panic if the primary value is nil.
func MakeSourceConfig(parent, name string, primary *SourceConfigValue, additional ...*SourceConfigValue) (sc *SourceConfig) {
	var parentPointer *string
	if parent != "" {
		parentPointer = &parent
	}
	var values ConfigSourceValues
	if parentPointer == nil && primary == nil {
		panic("MakeSourceConfig (without a parent) requires a primary value")
	} else if primary != nil {
		values = append(values, primary)
	}
	values = append(values, additional...)
	return &SourceConfig{
		Name:   name,
		Parent: parentPointer,
		Values: values,
	}
}
