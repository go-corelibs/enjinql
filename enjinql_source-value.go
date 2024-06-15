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

	"github.com/go-corelibs/go-sqlbuilder"
)

type sourceValueType uint8

const (
	gInvalidValue sourceValueType = iota
	gIntValue
	gBoolValue
	gFloatValue
	gStringValue
	gLinkValue
	gTimeValue
)

type cSourceValue struct {
	ivt sourceValueType
	key string
	opt *sqlbuilder.ColumnOption
}

func (c cSourceValue) columnConfig() (column sqlbuilder.ColumnConfig, err error) {
	switch c.ivt {
	case gIntValue, gLinkValue:
		column = sqlbuilder.IntColumn(c.key, c.opt)
	case gBoolValue:
		column = sqlbuilder.BoolColumn(c.key, c.opt)
	case gFloatValue:
		column = sqlbuilder.FloatColumn(c.key, c.opt)
	case gStringValue:
		column = sqlbuilder.StringColumn(c.key, c.opt)
	case gTimeValue:
		column = sqlbuilder.DateColumn(c.key, c.opt)
	default:
		err = fmt.Errorf("invalid source values type for: %q", c.key)
	}
	return
}
