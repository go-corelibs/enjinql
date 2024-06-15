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
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/go-corelibs/context"
)

func (esh *cEqlShell) renderSplash() (output string) {
	tw := table.NewWriter()
	tw.SuppressTrailingSpaces()
	tw.SetColumnConfigs([]table.ColumnConfig{
		{AutoMerge: true},
		{AutoMerge: true},
	})
	tw.SetTitle("EnjinQL Shell " + gShellVersion)
	tw.AppendRow(table.Row{"Dialect", esh.eql.SqlDialect().Name()})
	config := esh.eql.Config()
	if config.Prefix == "" {
		tw.AppendRow(table.Row{"Prefix", "(nil)"})
	} else {
		tw.AppendRow(table.Row{"Prefix", config.Prefix})
	}
	if names := config.Sources.DataNames(); len(names) > 0 {
		tw.AppendRow(table.Row{"Data Sources", strings.Join(names, ", ")})
	}
	if names := config.Sources.LinkNames(); len(names) > 0 {
		tw.AppendRow(table.Row{"Link Sources", strings.Join(names, ", ")})
	}
	if names := config.Sources.JoinNames(); len(names) > 0 {
		tw.AppendRow(table.Row{"Join Sources", strings.Join(names, ", ")})
	}
	output += "\n"
	output += tw.Render() + "\n"
	output += `(type "help" for usage information)`
	output += "\n"
	return
}

func (esh *cEqlShell) renderSources() (output string) {
	tw := table.NewWriter()
	tw.SuppressTrailingSpaces()
	tw.SetColumnConfigs([]table.ColumnConfig{
		{},
		{WidthMax: 10, WidthMaxEnforcer: text.WrapText},
		{WidthMax: 10, WidthMaxEnforcer: text.WrapText},
		{WidthMax: 10, WidthMaxEnforcer: text.WrapText},
	})
	tw.AppendHeader(table.Row{"type", "name", "parent", "values"}, table.RowConfig{AutoMerge: true})
	config := esh.eql.Config()
	for _, sc := range config.Sources {
		var parent string
		if sc.Parent != nil {
			parent = *sc.Parent
		} else {
			parent = "-"
		}
		tw.AppendRow(table.Row{
			sc.Type().String(),
			sc.Name,
			parent,
			strings.Join(sc.Values.Names(), ", "),
		})
	}
	output += "\n"
	output += tw.Render() + "\n"
	return
}

func (esh *cEqlShell) renderSQL(parsed, query string, argv []interface{}) (output string) {
	tw := table.NewWriter()
	tw.SuppressTrailingSpaces()
	tw.SetColumnConfigs([]table.ColumnConfig{
		{},
		{WidthMax: 10, WidthMaxEnforcer: text.WrapText},
		{WidthMax: 10, WidthMaxEnforcer: text.WrapText},
		{WidthMax: 10, WidthMaxEnforcer: text.WrapText},
	})
	tw.AppendRow(table.Row{"EQL", parsed}, table.RowConfig{AutoMerge: true})
	tw.AppendRow(table.Row{"SQL", query}, table.RowConfig{AutoMerge: true})
	tw.AppendRow(table.Row{"ARG", fmt.Sprintf("%v", argv)}, table.RowConfig{AutoMerge: true})
	output += "\n"
	output += tw.Render() + "\n"
	return
}

func (esh *cEqlShell) renderResults(columns []string, results context.Contexts) (output string) {
	count := len(results)

	if count > 0 {
		tw := table.NewWriter()
		tw.SuppressTrailingSpaces()

		header := table.Row{"#"}
		for _, column := range columns {
			header = append(header, column)
		}
		tw.AppendHeader(header)

		for idx, result := range results {
			row := table.Row{idx + 1}
			for _, key := range columns {
				row = append(row, result[key])
			}
			tw.AppendRow(row)
		}

		output += tw.Render() + "\n"
	}
	return
}
