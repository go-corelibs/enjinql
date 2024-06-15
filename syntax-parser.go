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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/participle/v2/lexer"

	clStrings "github.com/go-corelibs/strings"
)

/*

 #       +-----------+-------------------------------------> key selection
 #       |           |                 +-------------------> source
 #       |           |                 |             +-----> condition key
 #       v           v                 v             v
 LOOKUP words.word, words.flat WITHIN word_letters.letter = "a"
 OFFSET 10 LIMIT 10

 * join "words" and "word_letters" sources
 * with word_letters condition
 * pick out words.word and words.flat

WITHIN <condition> AND <condition> OR <condition>

*/

const (
	glInt            = `\b(\d+)\b`
	glFloat          = `\b(\d*\.\d+)\b`
	glIdent          = `\b([_a-zA-Z][_a-zA-Z0-9]*)\b`
	glOperator       = `(==|\!=|\^=|\$=|\~=|\*=|<=|>=|<>|<|>)`
	glEmptySpace     = `\s+`
	glPlaceholder    = `\{\d+\}`
	glPunctuation    = `[.,;!()]`
	glSingleQuoted   = `'(?:\\'|[^'])*'`
	glDoubleQuoted   = `"(?:\\"|[^"])*"`
	glBacktickQuoted = "`(?:\\\\`|[^`])*`"
)

var (
	gLexerKeywords = []string{
		"DISTINCT",
		"LOOKUP", "OFFSET", "WITHIN", "RANDOM",
		"QUERY", "COUNT", "FALSE", "ORDER", "LIMIT",
		"DESC", "LIKE", "TRUE", "NULL",
		"AND", "ASC", "DSC", "NOT", "NIL",
		"AS", "BY", "IN", "OR",
		"SW", "EW", "CS", "CF",
	}
	gSyntaxLexer = lexer.MustSimple([]lexer.SimpleRule{
		{Name: `Placeholder`, Pattern: glPlaceholder},
		{Name: `Int`, Pattern: glInt},
		{Name: `Float`, Pattern: glFloat},
		{Name: `String`, Pattern: `(` + strings.Join([]string{glSingleQuoted, glDoubleQuoted, glBacktickQuoted}, "|") + `)`},
		{Name: `Operator`, Pattern: glOperator},
		{Name: `Punctuation`, Pattern: glPunctuation},
		{Name: `Keyword`, Pattern: `(?i)\b(` + strings.Join(gLexerKeywords, "|") + `)\b`},
		{Name: `Ident`, Pattern: glIdent},
		{Name: `whitespace`, Pattern: glEmptySpace},
	})
)

// GetSyntaxEBNF returns the EBNF text representing the Enjin Query Language
func GetSyntaxEBNF() (ebnf string) {
	return gSyntaxParser.String()
}

// GetLexerJSON returns a JSON representation of the syntax lexer
func GetLexerJSON() (text string) {
	data, _ := json.MarshalIndent(gSyntaxLexer, "", "  ")
	return string(data)
}

// ParseSyntax parses the input string and returns an initialized Syntax tree
func ParseSyntax[V []byte | string](input V) (parsed *Syntax, err error) {
	switch t := interface{}(input).(type) {
	case []byte:
		parsed, err = gSyntaxParser.ParseBytes("enjinql", t)
	case string:
		parsed, err = gSyntaxParser.ParseString("enjinql", t) //, participle.Trace(os.Stdout))
	}
	if parsed != nil && err == nil {
		if err = parsed.init(); err != nil {
		} else if err = parsed.Validate(); err != nil {
			parsed = nil
		}
	}
	return
}

var (
	txPlaceholders = regexp.MustCompile(`(\{\d+\})`)
)

func parsePlaceholder(input string) (pos int, ok bool) {
	if size := len(input); size >= 3 {
		var err error
		pos, err = strconv.Atoi(input[1 : size-1])
		ok = err == nil
	}
	return
}

func scanPlaceholders(input string) (placeholders []string) {
	m := txPlaceholders.FindAllStringSubmatch(input, -1)
	for _, mm := range m {
		placeholders = append(placeholders, mm[1])
	}
	return
}

func rplPlaceholders(input string, argc int, argv []interface{}) string {
	for _, placeholder := range scanPlaceholders(input) {
		if pos, ok := parsePlaceholder(placeholder); ok {
			if pos > 0 && pos <= argc {
				if _, ok := argv[pos-1].(string); ok {
					input = strings.Replace(input, placeholder, "%["+strconv.Itoa(pos)+"]q", 1)
				} else if _, ok := argv[pos-1].(time.Time); ok {
					input = strings.Replace(input, placeholder, "%["+strconv.Itoa(pos)+"]q", 1)
				} else {
					input = strings.Replace(input, placeholder, "%["+strconv.Itoa(pos)+"]v", 1)
				}
			}
		}
	}
	return input
}

func PrepareSyntax(format string, argv ...interface{}) (prepared string, err error) {
	if prepared = format; len(argv) == 0 {
		return
	}
	argc := len(argv)

	// convert all {\d} placeholders, that are not within quoted strings, with either %[\d]q (string, time) or %[\d]v

	var modified string
	for remainder := format; remainder != ""; {
		if before, quoted, after, found := clStrings.ScanQuote(remainder); found {
			modified += rplPlaceholders(before, argc, argv)
			modified += strconv.Quote(quoted)
			remainder = after
		} else {
			modified += rplPlaceholders(before, argc, argv)
			break
		}
	}

	// process fmt placeholders
	prepared = fmt.Sprintf(modified, argv...)
	return
}
