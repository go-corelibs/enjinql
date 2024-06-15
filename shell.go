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
	"strings"
	"time"

	"github.com/abiosoft/ishell/v2"

	"github.com/go-corelibs/context"
)

var (
	gShellVersion = "v0.1.0"
)

// Shell is a simple interface for managing an interactive eql shell session
type Shell interface {
	// Run starts the interactive shell
	Run()
	// Stop stops the interactive shell
	Stop()
	// Close shuts down the shell completely
	Close()
	// Process runs shell using arguments in non-interactive mode
	Process(argv ...string) (err error)
}

type cEqlShell struct {
	eql   EnjinQL
	shell *ishell.Shell
}

// NewShell starts a new EnjinQL interactive shell, creating a new default
// shell configuration if the shell argument is nil
func NewShell(eql EnjinQL, shell *ishell.Shell) Shell {

	if shell == nil {
		shell = ishell.New()
		shell.SetPrompt("eql> ")
		shell.SetHomeHistoryPath(".enjinql_history")
		shell.IgnoreCase(true)
		//shell.SetPager("less", []string{"-SR"})

		shell.EOF(func(c *ishell.Context) {
			c.Printf("# exiting now\n")
			c.Stop()
		})

		shell.Interrupt(func(c *ishell.Context, count int, input string) {
			if count > 1 {
				c.Stop()
				return
			}
			c.Printf("# press <CTRL+c> one more time to exit\n")
		})
	}

	esh := &cEqlShell{eql: eql, shell: shell}

	shell.Println(esh.renderSplash())

	shell.AddCmd(&ishell.Cmd{
		Name:     "lookup",
		Help:     "LOOKUP <statement>",
		LongHelp: "perform an EQL LOOKUP statement",
		Func:     esh.cmdLookup,
	})

	shell.AddCmd(&ishell.Cmd{
		Name:     "select",
		Help:     "SELECT <query>",
		LongHelp: "perform an SQL SELECT statement",
		Func:     esh.cmdSelect,
	})

	shell.AddCmd(&ishell.Cmd{
		Name:     "plan",
		Help:     "PLAN <LOOKUP|QUERY> <statement>",
		LongHelp: "display the SQL table join plan for an EQL statement",
		Func:     esh.cmdPlan,
	})

	shell.AddCmd(&ishell.Cmd{
		Name:     "show",
		Help:     "SHOW <LOOKUP|QUERY> <statement>",
		LongHelp: "display the SQL query and arguments for an EQL statement",
		Func:     esh.cmdShow,
	})

	shell.AddCmd(&ishell.Cmd{
		Name:     "explain",
		Help:     "EXPLAIN <LOOKUP|QUERY> <statement>",
		LongHelp: "explain the SQL query statement for an EQL statement",
		Func:     esh.cmdExplain,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "info",
		Help: "display a summary of enjinql sources",
		Func: esh.cmdSourceInfo,
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "config",
		Help: "display the complete enjinql config (json)",
		Func: esh.cmdConfig,
	})

	return esh
}

func (esh *cEqlShell) Run() {
	esh.shell.Run()
}

func (esh *cEqlShell) Stop() {
	esh.shell.Stop()
}

func (esh *cEqlShell) Close() {
	esh.shell.Close()
}

func (esh *cEqlShell) Process(argv ...string) (err error) {
	return esh.shell.Process(argv...)
}

func (esh *cEqlShell) cmdConfig(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)
	_ = c.ShowPaged(esh.eql.String())
}

func (esh *cEqlShell) cmdSourceInfo(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)
	c.Println(esh.renderSources())
}

func (esh *cEqlShell) cmdLookup(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	query := strings.Join(c.RawArgs, " ")

	var ee error
	var columns []string
	var results context.Contexts

	start := time.Now()
	if columns, results, ee = esh.eql.Perform(query); ee != nil {
		c.Printf("error: %v\n", ee)
		return
	}
	delta := time.Now().Sub(start)

	c.Print("\n" + esh.renderResults(columns, results))
	c.Printf("# %d results in %v\n\n", len(results), delta)
}

func (esh *cEqlShell) cmdSelect(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	query := strings.Join(c.RawArgs, " ")

	var ee error
	var columns []string
	var results context.Contexts

	start := time.Now()

	if columns, results, ee = esh.eql.SqlQuery(query); ee != nil {
		c.Printf("error: %v\n", ee)
		return
	}

	delta := time.Now().Sub(start)

	c.Print("\n" + esh.renderResults(columns, results))
	c.Printf("# %d results in %v\n\n", len(results), delta)
}

func (esh *cEqlShell) cmdPlan(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	statement := strings.Join(c.RawArgs[1:], " ")

	start := time.Now()
	var ee error
	var _, verbose string
	if _, verbose, ee = esh.eql.Plan(statement); ee != nil {
		c.Printf("error: %v\n", ee)
		return
	}
	delta := time.Now().Sub(start)

	c.Print("\n")

	lines := strings.Split(verbose, "\n")
	for _, line := range lines {
		c.Print("\t" + line + "\n")
	}

	c.Printf("# prepared in %v\n\n", delta)
}

func (esh *cEqlShell) cmdShow(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	statement := strings.Join(c.RawArgs[1:], " ")

	start := time.Now()
	var ee error
	var parsed *Syntax
	var query string
	var argv []interface{}
	if parsed, ee = esh.eql.Parse(statement); ee != nil {
		c.Printf("error: %v\n", ee)
		return
	} else if query, argv, ee = esh.eql.ParsedToSql(parsed); ee != nil {
		c.Printf("error: %v\n", ee)
		return
	}
	delta := time.Now().Sub(start)

	c.Print("\n" + esh.renderSQL(parsed.String(), query, argv))
	c.Printf("# prepared in %v\n\n", delta)
}

func (esh *cEqlShell) cmdExplain(c *ishell.Context) {
	c.ShowPrompt(false)
	defer c.ShowPrompt(true)

	statement := strings.Join(c.RawArgs[1:], " ")

	start := time.Now()
	var ee error
	var columns []string
	var results context.Contexts
	var query string
	var argv []interface{}
	if query, argv, ee = esh.eql.ToSQL(statement); ee != nil {
		c.Printf("error: %v\n", ee)
		return
	} else if columns, results, ee = esh.eql.SqlQuery("EXPLAIN "+query, argv...); ee != nil {
		c.Printf("error: %v\n", ee)
		return
	}
	delta := time.Now().Sub(start)

	c.Print("\n" + esh.renderResults(columns, results))
	c.Printf("# prepared in %v\n\n", delta)
}
