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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
)

// Config is the structure for configuring a New EnjinQL instance
//
// Config structures can be constructed manually, simply instantiate the Go
// types and build the structure directly.
// To check for errors, call the Config.Validate method.
//
// Another way to create Config structures is with JSON and using ParseConfig
// to both unmarshal and validate the resulting Config instance.
//
// The last way is to use the builder methods in a long chain to build the
// Config programmatically
//
// For example, to recreate a config with the default PageSource and others
// for the page titles and descriptions:
//
//	bePage, err := NewConfig("be_eql").    // start building a Config
//	    NewSource("page").                 // start building the SourceConfig
//	    AddStringValue("shasum", 10).      // add shasum column
//	    AddStringValue("language", 10).    // add language code column
//	    AddStringValue("type", 48).        // add page type column
//	    AddStringValue("url", 1024).       // add page URL path column
//	    AddStringValue("stub", -1).        // add page stub column
//	    DoneSource().                      // done making this particular source
//	    NewSource("extra").                // start another SourceConfig
//	    AddStringValue("title", 200).      // add the title column
//	    AddStringValue("description", -1). // add the description
//	    DoneSource().                      // done making this particular source
//	    Make()
type Config struct {
	Prefix  string        `json:"prefix,omitempty"`
	Sources ConfigSources `json:"sources,omitempty"`
}

// ParseConfig unmarshalls the given JSON data into a new Config instance
func ParseConfig[V string | []byte](data V) (c *Config, err error) {
	config := &Config{}
	if err = json.Unmarshal([]byte(data), config); err == nil {
		if err = config.Validate(); err == nil {
			c = config
		}
		return
	}
	err = fmt.Errorf("%w: %w", ErrInvalidJSON, err)
	return
}

// NewConfig returns a new Config instance with only the given prefix set.
// All prefix values given are combined into a single snake_cased prefix string
func NewConfig(prefix ...string) (c *Config) {
	return &Config{Prefix: strcase.ToSnake(strings.Join(prefix, "_"))}
}

func (c *Config) Clone() (cloned *Config) {
	cloned = &Config{
		Prefix:  c.Prefix,
		Sources: c.Sources.Clone(),
	}
	cloned.Sources.update(cloned)
	return
}

// Serialize is a convenience method for returning (unindented) JSON data
// representing this Config instance, use ParseConfig to restore the Config
func (c *Config) Serialize() (output string) {
	if data, err := json.Marshal(c); err == nil {
		output = string(data)
	}
	return
}

// String returns (indented) JSON data representing this Config instance, use
// ParseConfig to restore the Config
func (c *Config) String() (output string) {
	if data, err := json.MarshalIndent(c, "", "\t"); err == nil {
		output = string(data)
	}
	return
}

// Validate checks the Config instance for any errors, returning the first one
// found
func (c *Config) Validate() (err error) {

	c.Sources.update(c)

	for _, validator := range gConfigValidators.configs {
		if err = validator.fn(c); err != nil {
			return
		}
	}

	dataSources := make(map[string]struct{})
	for idx, sc := range c.Sources {

		for _, validator := range gConfigValidators.sources {
			if err = validator.fn(c, idx, sc); err != nil {
				return
			}
		}

		for jdx, scv := range sc.Values {
			for _, validator := range gConfigValidators.values {
				if err = validator.fn(c, idx, sc, jdx, scv); err != nil {
					return
				}
			}
		}

		if sc.Parent != nil {
			if _, present := dataSources[*sc.Parent]; !present {
				return fmt.Errorf("%w: %w (%q needs %q declared first)", ErrInvalidConfig, ErrParentNotFound, sc.Name, *sc.Parent)
			}
		}

		dataSources[sc.Name] = struct{}{}
	}
	return
}

func (c *Config) Make() (config *Config, err error) {
	if err = c.Validate(); err == nil {
		config = c
	}
	return
}

func (c *Config) AddSource(source *SourceConfig) *Config {
	c.Sources = append(c.Sources, source)
	return c
}

func (c *Config) NewSource(name string) (source *SourceConfig) {
	source = &SourceConfig{
		Name:   name,
		config: c,
	}
	c.AddSource(source)
	return
}
