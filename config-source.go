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
	"github.com/go-corelibs/slices"
	"github.com/go-corelibs/values"
)

// SourceConfig is the structure for configuring a specific source
type SourceConfig struct {
	Name   string             `json:"name"`
	Parent *string            `json:"parent,omitempty"`
	Values ConfigSourceValues `json:"values"`
	Unique [][]string         `json:"unique,omitempty"`
	Index  [][]string         `json:"index,omitempty"`

	config *Config
}

func (sc *SourceConfig) update(c *Config) {
	sc.config = c
	sc.Values.update(c)
}

func (sc *SourceConfig) Clone() (cloned *SourceConfig) {
	var parent *string
	if sc.Parent != nil {
		parent = values.Ref(*sc.Parent)
	}
	cloned = &SourceConfig{
		Name:   sc.Name,
		Parent: parent,
		Values: sc.Values.Clone(),
		Unique: slices.Copy(sc.Unique),
		Index:  slices.Copy(sc.Index),
	}
	return
}

func (sc *SourceConfig) Type() (t SourceConfigType) {
	switch {
	case sc.Parent == nil && !sc.Values.HasLinks():
		return DataSourceType
	case sc.Parent != nil && sc.Values.HasLinks():
		return JoinSourceType
	default: // case sc.Parent != nil || sc.Values.HasLinks():
		return LinkSourceType
	}
}

// SetParent configures the SourceConfig.Parent setting
func (sc *SourceConfig) SetParent(name string) *SourceConfig {
	sc.Parent = &name
	return sc
}

// AddValue adds the given SourceConfigValue
func (sc *SourceConfig) AddValue(v *SourceConfigValue) *SourceConfig {
	v.update(sc.config)
	sc.Values = append(sc.Values, v)
	return sc
}

func (sc *SourceConfig) NewIntValue(key string) *SourceConfig {
	sc.AddValue(&SourceConfigValue{
		Int: &SourceConfigValueInt{
			Key: key,
		},
	})
	return sc
}

func (sc *SourceConfig) NewBoolValue(key string) *SourceConfig {
	sc.AddValue(&SourceConfigValue{
		Bool: &SourceConfigValueBool{
			Key: key,
		},
	})
	return sc
}

func (sc *SourceConfig) NewFloatValue(key string) *SourceConfig {
	sc.AddValue(&SourceConfigValue{
		Float: &SourceConfigValueFloat{
			Key: key,
		},
	})
	return sc
}

// NewStringValue adds a string value column with the given size.
// If the size is less-than or equal-to zero, the column will be some sort of
// TEXT type depending on the specific SQL service used
func (sc *SourceConfig) NewStringValue(key string, size int) *SourceConfig {
	sc.AddValue(&SourceConfigValue{
		String: &SourceConfigValueString{
			Key:  key,
			Size: size,
		},
	})
	return sc
}

// NewLinkedValue adds a cross-table link to another source
func (sc *SourceConfig) NewLinkedValue(table, key string) *SourceConfig {
	sc.AddValue(&SourceConfigValue{
		Linked: &SourceConfigValueLinked{
			Source: table,
			Key:    key,
		},
	})
	return sc
}

// AddUnique add the given keys to the list of unique constraints
func (sc *SourceConfig) AddUnique(keys ...string) *SourceConfig {
	sc.Unique = append(sc.Unique, keys)
	return sc
}

// AddIndex adds the given keys to the list of indexes
func (sc *SourceConfig) AddIndex(keys ...string) *SourceConfig {
	sc.Index = append(sc.Index, keys)
	return sc
}

// DoneSource completes this SourceConfig builder chain
func (sc *SourceConfig) DoneSource() *Config {
	return sc.config
}
