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
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v2"

	"github.com/go-corelibs/enjinql"
)

var (
	Version = "v0.1.0"
	gApp    = cli.App{
		Name:      "enjinql",
		Usage:     "enjin query language command line interface",
		UsageText: "enjinql [global options] <command> [options] [arguments...]",
		Version:   Version,
		Commands: []*cli.Command{
			{
				Name:        "ebnf",
				Description: "print an EBNF representation of the enjinql syntax",
				Action: func(ctx *cli.Context) (err error) {
					_, _ = fmt.Fprintf(os.Stdout, enjinql.GetSyntaxEBNF()+"\n")
					return
				},
			},
			{
				Name:        "lexer",
				Description: "print a JSON representation of the enjinql syntax lexer",
				Action: func(ctx *cli.Context) (err error) {
					_, _ = fmt.Fprintf(os.Stdout, enjinql.GetLexerJSON()+"\n")
					return
				},
			},
			{
				Name:        "shell",
				Description: "start an EQL shell session",
				Action:      actionShellFn,
				Flags: []cli.Flag{
					&cli.PathFlag{
						Name:      "config",
						Usage:     "EnjinQL config file",
						Required:  true,
						TakesFile: true,
						Aliases:   []string{"c"},
					},
					&cli.StringFlag{
						Name:     "dsn",
						Usage:    "database connection string",
						Required: true,
						Aliases:  []string{"d"},
					},
				},
			},
		},
	}
)

func main() {
	if err := gApp.Run(os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
