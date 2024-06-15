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

package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/enjinql"
	"github.com/go-corelibs/go-sqlbuilder"
	"github.com/go-corelibs/go-sqlbuilder/dialects"
	"github.com/go-corelibs/path"
)

func setupEQL(ctx *cli.Context) (eql enjinql.EnjinQL, err error) {

	var data []byte
	var cfg, dsn string
	var config *enjinql.Config

	if !ctx.IsSet("config") {
		err = fmt.Errorf("--config is required")
		return
	} else if !ctx.IsSet("dsn") {
		err = fmt.Errorf("--dsn is required")
		return
	}

	cfg = ctx.String("config")
	dsn = ctx.String("dsn")

	if !path.IsFile(cfg) {
		err = fmt.Errorf("--config is not a file: %q", cfg)
		return
	} else if data, err = os.ReadFile(cfg); err != nil {
		err = fmt.Errorf("error reading --config: %q - %v", cfg, err)
		return
	} else if config, err = enjinql.ParseConfig(data); err != nil {
		err = fmt.Errorf("error parsing --config: %q - %v", cfg, err)
		return
	}

	var dbh *sql.DB
	var dialect sqlbuilder.Dialect

	switch {
	case strings.HasPrefix(dsn, "sqlite://"):
		dsn = dsn[9:]
		dialect = dialects.Sqlite{}
		if dbh, err = sql.Open("sqlite3", dsn); err != nil {
			err = fmt.Errorf("error connecting to sqlite: %v", err)
			return
		}
	default:
		err = fmt.Errorf("only sqlite supported at this time")
		return
	}

	if eql, err = enjinql.New(config, dbh, dialect); err != nil {
		err = fmt.Errorf("error making enjinql instance: %v", err)
	}
	return
}
