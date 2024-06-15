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

// ConfigSources is a slice type for providing Config builder methods
type ConfigSources []*SourceConfig

func (s ConfigSources) update(c *Config) {
	for _, source := range s {
		source.update(c)
	}
}

func (s ConfigSources) Clone() (cloned ConfigSources) {
	for _, sc := range s {
		cloned = append(cloned, sc.Clone())
	}
	return
}

// Get returns a clone of the named SourceConfig
func (s ConfigSources) Get(name string) (cloned *SourceConfig) {
	for _, sc := range s {
		if sc.Name == name {
			cloned = sc.Clone()
			return
		}
	}
	return
}

// Names returns a list of all source names, in the order they were added
func (s ConfigSources) Names() (names []string) {
	for _, sc := range s {
		names = append(names, sc.Name)
	}
	return
}

// DataNames returns a list of all data source names (sources with no parent),
// in the order they were added
func (s ConfigSources) DataNames() (names []string) {
	for _, sc := range s {
		if sc.Type() == DataSourceType {
			names = append(names, sc.Name)
		}
	}
	return
}

// LinkNames returns a list of all link source names (sources with a parent
// or has at least one value linked), in the order they were added
func (s ConfigSources) LinkNames() (names []string) {
	for _, sc := range s {
		if sc.Type() == LinkSourceType {
			names = append(names, sc.Name)
		}
	}
	return
}

// JoinNames returns a list of all link source names (sources with a parent
// and has at least one value linked), in the order they were added
func (s ConfigSources) JoinNames() (names []string) {
	for _, sc := range s {
		if sc.Type() == JoinSourceType {
			names = append(names, sc.Name)
		}
	}
	return
}
