// Copyright (c) 2024  The Go-CoreLibs Authors
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

const (
	PageSource       = "page"
	PageSourceIdKey  = "page_id"
	PageShasumKey    = "shasum"
	PageLanguageKey  = "language"
	PageTypeKey      = "type"
	PageArchetypeKey = "archetype"
	PageCreatedKey   = "created"
	PageUpdatedKey   = "updated"
	PageUrlKey       = "url"
	PageStubKey      = "stub"
)

const (
	// MaxUrlPathSize is the recommended 2000-character limit on overall URL
	// size, minus 256 for the domain segment and minus another nine for the
	// https://
	MaxUrlPathSize = 2000 - 256 - 8

	// MaxPageTypeSize is the Go-Enjin recommended 64-character limit on the
	// total length of custom page type and archetype names
	MaxPageTypeSize = 64

	// PageShasumSize is the length of a Go-Enjin page "shasum" identifier,
	// which is the first 10-characters of the SHA-256 hash of the complete
	// page content (front-matter plus body)
	PageShasumSize = 10
)

// PageSourceConfig returns a new SourceConfig, preset with the primary source
// settings required for the Go-Enjin project. The page SourceConfig is preset
// with five columns in addition to the default id column present with all
// sources:
//
//	+------+-----------+-------------------------------------------+
//	| size | column    | description                               |
//	+------+-----------+-------------------------------------------+
//	|  10  | shasum    | a unique identifier used within Go-Enjin  |
//	|  10  | language  | the language code for a page              |
//	|  64  | type      | the type of page                          |
//	|  64  | archetype | the page archetype                        |
//	| 2000 | url       | the URL path (absolute) to a page         |
//	|  -1  | stub      | JSON context for filesystem lookup        |
//	+------+-----------+-------------------------------------------+
func PageSourceConfig() (sc *SourceConfig) {
	return MakeSourceConfig(
		"",
		PageSource,
		NewStringValue(PageShasumKey, PageShasumSize),
		NewStringValue(PageLanguageKey, 10),
		NewStringValue(PageTypeKey, MaxPageTypeSize),
		NewStringValue(PageArchetypeKey, MaxPageTypeSize),
		NewTimeValue(PageCreatedKey),
		NewTimeValue(PageUpdatedKey),
		NewStringValue(PageUrlKey, MaxUrlPathSize),
		NewStringValue(PageStubKey, -1),
	).
		AddUnique(PageShasumKey).
		AddUnique(PageShasumKey, PageUrlKey).
		// single column indexes
		AddIndex(PageUrlKey).
		AddIndex(PageTypeKey).
		AddIndex(PageShasumKey).
		AddIndex(PageCreatedKey).
		AddIndex(PageUpdatedKey).
		AddIndex(PageLanguageKey).
		// two column indexes
		AddIndex(PageUrlKey, PageShasumKey).AddIndex(PageShasumKey, PageUrlKey).
		AddIndex(PageTypeKey, PageShasumKey).AddIndex(PageShasumKey, PageTypeKey).
		AddIndex(PageCreatedKey, PageShasumKey).AddIndex(PageShasumKey, PageCreatedKey).
		AddIndex(PageUpdatedKey, PageShasumKey).AddIndex(PageShasumKey, PageUpdatedKey).
		AddIndex(PageLanguageKey, PageShasumKey).AddIndex(PageShasumKey, PageLanguageKey).
		AddIndex(PageArchetypeKey, PageShasumKey).AddIndex(PageShasumKey, PageArchetypeKey)
}
