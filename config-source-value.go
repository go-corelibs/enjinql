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

// SourceConfigValue is the structure for configuring a specific value indexed
// by the SourceConfig
type SourceConfigValue struct {
	Int    *SourceConfigValueInt    `json:"int,omitempty"`
	Bool   *SourceConfigValueBool   `json:"bool,omitempty"`
	Time   *SourceConfigValueTime   `json:"time,omitempty"`
	Float  *SourceConfigValueFloat  `json:"float,omitempty"`
	String *SourceConfigValueString `json:"string,omitempty"`
	Linked *SourceConfigValueLinked `json:"linked,omitempty"`

	config *Config
}

// NewIntValue is a convenience wrapper to construct an integer SourceConfigValue
func NewIntValue(key string) *SourceConfigValue {
	return &SourceConfigValue{Int: &SourceConfigValueInt{Key: key}}
}

// NewBoolValue is a convenience wrapper to construct a boolean SourceConfigValue
func NewBoolValue(key string) *SourceConfigValue {
	return &SourceConfigValue{Bool: &SourceConfigValueBool{Key: key}}
}

// NewTimeValue is a convenience wrapper to construct a boolean SourceConfigValue
func NewTimeValue(key string) *SourceConfigValue {
	return &SourceConfigValue{Time: &SourceConfigValueTime{Key: key}}
}

// NewFloatValue is a convenience wrapper to construct a decimal SourceConfigValue
func NewFloatValue(key string) *SourceConfigValue {
	return &SourceConfigValue{Float: &SourceConfigValueFloat{Key: key}}
}

// NewStringValue is a convenience wrapper to construct a string SourceConfigValue
func NewStringValue(key string, size int) *SourceConfigValue {
	return &SourceConfigValue{String: &SourceConfigValueString{Key: key, Size: size}}
}

// NewLinkedValue is a convenience wrapper to construct a linked SourceConfigValue
func NewLinkedValue(source, key string) *SourceConfigValue {
	return &SourceConfigValue{Linked: &SourceConfigValueLinked{Source: source, Key: key}}
}

func (scv *SourceConfigValue) update(c *Config) {
	scv.config = c
	switch {
	case scv.Int != nil:
		scv.Int.config = c
	case scv.Bool != nil:
		scv.Bool.config = c
	case scv.Time != nil:
		scv.Time.config = c
	case scv.Float != nil:
		scv.Float.config = c
	case scv.String != nil:
		scv.String.config = c
	case scv.Linked != nil:
		scv.Linked.config = c
	}
}

func (scv *SourceConfigValue) Clone() (cloned *SourceConfigValue) {
	switch {
	case scv.Int != nil:
		return &SourceConfigValue{Int: &SourceConfigValueInt{
			Key: scv.Int.Key,
		}}
	case scv.Bool != nil:
		return &SourceConfigValue{Bool: &SourceConfigValueBool{
			Key: scv.Bool.Key,
		}}
	case scv.Time != nil:
		return &SourceConfigValue{Time: &SourceConfigValueTime{
			Key: scv.Time.Key,
		}}
	case scv.Float != nil:
		return &SourceConfigValue{Float: &SourceConfigValueFloat{
			Key: scv.Float.Key,
		}}
	case scv.String != nil:
		return &SourceConfigValue{String: &SourceConfigValueString{
			Key:  scv.String.Key,
			Size: scv.String.Size,
		}}
	case scv.Linked != nil:
		return &SourceConfigValue{Linked: &SourceConfigValueLinked{
			Source: scv.Linked.Source,
			Key:    scv.Linked.Key,
		}}
	}
	return
}

func (scv *SourceConfigValue) Name() (output string) {
	switch {
	case scv.Int != nil:
		return scv.Int.Key
	case scv.Bool != nil:
		return scv.Bool.Key
	case scv.Time != nil:
		return scv.Time.Key
	case scv.Float != nil:
		return scv.Float.Key
	case scv.String != nil:
		return scv.String.Key
	case scv.Linked != nil:
		return scv.Linked.Source + "_" + scv.Linked.Key
	}
	return
}

type SourceConfigValueInt struct {
	Key string `json:"key"`

	config *Config
}

type SourceConfigValueBool struct {
	Key string `json:"key"`

	config *Config
}

type SourceConfigValueTime struct {
	Key string `json:"key"`

	config *Config
}

type SourceConfigValueFloat struct {
	Key string `json:"key"`

	config *Config
}

type SourceConfigValueString struct {
	Key  string `json:"key"`
	Size int    `json:"size"`

	config *Config
}

type SourceConfigValueLinked struct {
	Source string `json:"table"`
	Key    string `json:"key"`

	config *Config
}
