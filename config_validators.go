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

	"github.com/iancoleman/strcase"
)

type cConfigValidator struct {
	label string
	fn    func(c *Config) (err error)
}

type cSourceConfigValidator struct {
	label string
	fn    func(c *Config, idx int, sc *SourceConfig) (err error)
}

type cSourceConfigValueValidator struct {
	label string
	fn    func(c *Config, idx int, sc *SourceConfig, jdx int, scv *SourceConfigValue) (err error)
}

var (
	gConfigValidators = struct {
		configs []cConfigValidator
		sources []cSourceConfigValidator
		values  []cSourceConfigValueValidator
	}{
		configs: []cConfigValidator{
			{
				"any prefix must be snake cased",
				func(c *Config) (err error) {
					if c.Prefix != "" {
						err = mustSnakeCase(c.Prefix)
					}
					return
				},
			},
			{
				"must have at least one source",
				func(c *Config) (err error) {
					if len(c.Sources) == 0 {
						return fmt.Errorf("%w: %w", ErrInvalidConfig, ErrNoSources)
					}
					return
				},
			},
		},
		sources: []cSourceConfigValidator{
			{
				"source config has a name",
				func(c *Config, idx int, sc *SourceConfig) (err error) {
					if sc.Name == "" {
						return fmt.Errorf("%w: %w (#%d)", ErrInvalidConfig, ErrUnnamedSource, idx+1)
					}
					return
				},
			},
			{
				"source config name is snake_cased",
				func(c *Config, idx int, sc *SourceConfig) (err error) {
					if err = mustSnakeCase(sc.Name); err != nil {
						return
					}
					return
				},
			},
			{
				"source config has at least one value",
				func(c *Config, idx int, sc *SourceConfig) (err error) {
					if len(sc.Values) == 0 && sc.Parent == nil {
						return fmt.Errorf("%w: %w (%q)", ErrInvalidConfig, ErrNoSourceValues, sc.Name)
					}
					return
				},
			},
			{
				"any parent set must be snake_cased",
				func(c *Config, idx int, sc *SourceConfig) (err error) {
					if sc.Parent != nil {
						err = mustSnakeCase(*sc.Parent)
					}
					return
				},
			},
		},
		values: []cSourceConfigValueValidator{
			{
				"value key present",
				func(c *Config, idx int, sc *SourceConfig, jdx int, scv *SourceConfigValue) (err error) {
					switch {
					case scv.Int != nil:
					case scv.Bool != nil:
					case scv.Time != nil:
					case scv.Float != nil:
					case scv.String != nil:
					case scv.Linked != nil:
					default:
						return fmt.Errorf("%w: %w (%q value #%d)", ErrInvalidConfig, ErrEmptySourceValue, sc.Name, jdx+1)
					}
					return
				},
			},
			{
				"value key is not empty and is snake_cased",
				func(c *Config, idx int, sc *SourceConfig, jdx int, scv *SourceConfigValue) (err error) {
					var key string
					switch {
					case scv.Int != nil:
						key = scv.Int.Key
					case scv.Bool != nil:
						key = scv.Bool.Key
					case scv.Time != nil:
						key = scv.Time.Key
					case scv.Float != nil:
						key = scv.Float.Key
					case scv.String != nil:
						key = scv.String.Key
					case scv.Linked != nil:
						if scv.Linked.Source == "" {
							return fmt.Errorf("%w: %w (%q value #%d)", ErrInvalidConfig, ErrEmptySourceValueKey, sc.Name, jdx+1)
						}
						key = scv.Linked.Key
					}
					if key == "" {
						return fmt.Errorf("%w: %w (%q value #%d)", ErrInvalidConfig, ErrEmptySourceValueKey, sc.Name, jdx+1)
					}
					return mustSnakeCase(key)
				},
			},
			{
				"linked table exists",
				func(c *Config, idx int, sc *SourceConfig, jdx int, scv *SourceConfigValue) (err error) {
					if scv.Linked != nil {
						var present bool
						for i := 0; i < idx; i++ {
							osc := c.Sources[i]
							if present = osc.Name == scv.Linked.Source; present {
								break
							}
						}
						if !present {
							return fmt.Errorf("%w: %w (%q needs %q declared first)", ErrInvalidConfig, ErrParentNotFound, sc.Name, scv.Linked.Source)
						}
					}
					return
				},
			},
		},
	}
)

func mustSnakeCase(input string) (err error) {
	if snake := strcase.ToSnake(input); snake != input {
		return fmt.Errorf("%w: %w (%q is not %q)", ErrInvalidConfig, ErrNotSnakeCased, input, snake)
	}
	return
}
