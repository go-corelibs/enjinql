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
	PagePermalinkSource   = "permalink"
	PagePermalinkShortKey = "short"
	PagePermalinkLongKey  = "long"
	ShortPermalinkSize    = 10 // size of [shasum.BriefLength]
	LongPermalinkSize     = 36 // length of hex-string uuid.V4 ((uuid.Size * 2) + 4)
)

func PagePermalinkSourceConfig() (sc *SourceConfig) {
	return MakeSourceConfig(
		PageSource,
		PagePermalinkSource,
		NewStringValue(PagePermalinkShortKey, ShortPermalinkSize),
		NewStringValue(PagePermalinkLongKey, LongPermalinkSize),
	).
		AddUnique(PagePermalinkShortKey).
		AddIndex(PagePermalinkLongKey).
		AddIndex(PagePermalinkShortKey).
		AddIndex(PagePermalinkLongKey, PagePermalinkShortKey).
		AddIndex(PagePermalinkShortKey, PagePermalinkLongKey)
}
