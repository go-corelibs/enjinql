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
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"

	. "github.com/smartystreets/goconvey/convey"

	clContext "github.com/go-corelibs/context"
	"github.com/go-corelibs/go-sqlbuilder"
	"github.com/go-corelibs/go-sqlbuilder/dialects"
	"github.com/go-corelibs/hrx"
	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/shasum"
	"github.com/go-corelibs/tdata"
	"github.com/go-corelibs/testdb"
	"github.com/go-corelibs/values"
)

func TestEnjinQL(t *testing.T) {
	td := tdata.New()

	sources := []*SourceConfig{
		{Name: "page", Parent: nil, Values: []*SourceConfigValue{
			{String: &SourceConfigValueString{Key: "shasum", Size: 10}},
		}, Unique: [][]string{{"shasum"}}, Index: [][]string{{"shasum"}}},
	}

	Convey("New", t, func() {

		Convey("with valid arguments", func() {

			tdb, err := testdb.NewTestDBWith(tdata.TempFile("", "enjinql.*.new.db"))
			SoMsg("sqlite db open error", err, ShouldBeNil)
			SoMsg("sqlite db instance", tdb, ShouldNotBeNil)
			defer tdb.Close()

			eql, err := New(&Config{Prefix: "be_eql", Sources: sources}, tdb.DBH(), dialects.Sqlite{})
			SoMsg("new enjinql error", err, ShouldBeNil)
			SoMsg("new enjinql struct", eql, ShouldNotBeNil)
			SoMsg("private truth", eql.private(nil), ShouldBeTrue)

			SoMsg("table check", tdb.Tables(), ShouldEqual, []string{"be_eql_page"})
			SoMsg("index check", tdb.Indexes(), ShouldEqual, []string{
				"be_eql_page_shasum",
			})

			columns, results, err := eql.Perform("LOOKUP .Shasum")
			SoMsg("parsed query error", err, ShouldBeNil)
			SoMsg("parsed query results", results, ShouldEqual, clContext.Contexts(nil))
			SoMsg("parsed query columns", len(columns), ShouldEqual, 0)

			SoMsg("unmarshal eql error", eql.Unmarshal([]byte("")), ShouldNotBeNil)

			data, err := eql.Marshal()
			SoMsg("marshal eql error", err, ShouldBeNil)
			expected := `{"prefix":"be_eql","sources":[{"name":"page","values":[{"string":{"key":"shasum","size":10}}],"unique":[["shasum"]],"index":[["shasum"]]}]}`
			SoMsg("marshal eql data", string(data), ShouldEqual, expected)

			indented := eql.String()
			expected = `{
	"prefix": "be_eql",
	"sources": [
		{
			"name": "page",
			"values": [
				{
					"string": {
						"key": "shasum",
						"size": 10
					}
				}
			],
			"unique": [
				[
					"shasum"
				]
			],
			"index": [
				[
					"shasum"
				]
			]
		}
	]
}`
			SoMsg("string eql data", indented, ShouldEqual, expected)

			Convey("prepareSQL", func() {

				specific, ok := eql.(*enjinql)
				SoMsg("specific *enjinql ok", ok, ShouldBeTrue)
				SoMsg("specific *enjinql", specific, ShouldNotBeNil)
				_, _, err = specific.prepareSQL(&Syntax{Query: true, Keys: []*SourceKey{{}}})
				SoMsg("prepareSQL syntax error", err, ShouldNotBeNil)

				_, _, err = specific.prepareSQL(&Syntax{Lookup: true, Keys: []*SourceKey{{Alias: values.Ref("nope")}}})
				SoMsg("prepareSQL syntax error", err, ShouldNotBeNil)

				//_, _, err = specific.prepareSQL(&Syntax{Lookup: true, Query: true, Keys: []*SourceKey{}})
				//SoMsg("prepareSQL syntax error", err, ShouldNotBeNil)

			})

		})

		Convey("with nil arguments", func() {

			tdb, err := testdb.NewTestDBWith(tdata.TempFile("", "enjinql.*.new.db"))
			SoMsg("sqlite db open error", err, ShouldBeNil)
			SoMsg("sqlite db instance", tdb, ShouldNotBeNil)
			defer tdb.Close()

			sc := PageSourceConfig()

			for _, test := range []struct {
				c       *Config
				dbh     *sql.DB
				dialect sqlbuilder.Dialect
			}{
				{nil, tdb.DBH(), dialects.Sqlite{}},
				{&Config{}, tdb.DBH(), dialects.Sqlite{}},
				{&Config{Sources: []*SourceConfig{sc}}, nil, dialects.Sqlite{}},
				{&Config{Sources: []*SourceConfig{sc}}, tdb.DBH(), nil},
			} {
				eql, err := New(test.c, test.dbh, test.dialect)
				So(err, ShouldNotBeNil)
				So(eql, ShouldBeNil)
			}

		})
	})

	Convey("Insert/Delete", t, func() {

		tdb, err := testdb.NewTestDBWith(tdata.TempFile("", "enjinql.*.new.db"))
		SoMsg("sqlite db open error", err, ShouldBeNil)
		SoMsg("sqlite db instance", tdb, ShouldNotBeNil)
		defer tdb.Close()

		config, err := NewConfig("be_eql").
			AddSource(PageSourceConfig()).
			Make()
		SoMsg("new config error", err, ShouldBeNil)
		SoMsg("new config instance", config, ShouldNotBeNil)
		SoMsg("new config valid", config.Validate(), ShouldBeNil)

		eql, err := New(config, tdb.DBH(), dialects.Sqlite{})
		SoMsg("new enjinql error", err, ShouldBeNil)
		SoMsg("new enjinql instance", eql, ShouldNotBeNil)

		now123, _ := time.Parse("2006-01-02 15:04", "1977-10-10 10:42")
		now012, _ := time.Parse("2006-01-02 15:04", "2017-02-14 21:34")
		now901, _ := time.Parse("2006-01-02 15:04", "2024-03-17 11:25")

		inserts := []struct {
			table  string
			id     Assertion
			err    Assertion
			values []interface{}
		}{
			{table: "page", id: ShouldNotBeZeroValue, err: ShouldBeNil, values: []interface{}{"1234567890", "en", "page", "", now123, now123, "/slug", `["stub"]`}},
			{table: "page", id: ShouldNotBeZeroValue, err: ShouldBeNil, values: []interface{}{"0123456789", "en", "page", "", now012, now012, "/other", `["other"]`}},
			{table: "nope", id: ShouldBeZeroValue, err: ShouldNotBeNil, values: []interface{}{}},
			{table: "page", id: ShouldBeZeroValue, err: ShouldNotBeNil, values: []interface{}{}},
			{table: "page", id: ShouldBeZeroValue, err: ShouldNotBeNil, values: []interface{}{10}},
			{table: "page", id: ShouldBeZeroValue, err: ShouldNotBeNil, values: []interface{}{"9012345678", "en", "page", "", now901, now901, "/another", `["another"]`, "too many"}},
		}

		tx, err := eql.SqlBegin()
		SoMsg("sql begin err", err, ShouldBeNil)
		SoMsg("sql transaction", tx, ShouldNotBeNil)
		stx := tx.TX()
		SoMsg("sql sub-tx", stx, ShouldNotBeNil)
		for idx, test := range inserts {
			id, err := stx.Insert(test.table, test.values...)
			SoMsg(fmt.Sprintf("insert test #%d error", idx), err, test.err)
			SoMsg(fmt.Sprintf("insert test #%d id", idx), id, test.id)
		}
		SoMsg("sql commit err", tx.Commit(), ShouldBeNil)

		columns, results, err := eql.Perform(`LOOKUP .ID, .Shasum ORDER BY .ID`)
		SoMsg("shasum[0] lookup error", err, ShouldBeNil)
		SoMsg("shasum[0] results length", len(results), ShouldEqual, 2)
		SoMsg("shasum[0] results values", results, ShouldEqual, clContext.Contexts{
			{"id": int64(1), "shasum": "1234567890"},
			{"id": int64(2), "shasum": "0123456789"},
		})
		SoMsg("shasum[0] results columns", len(columns), ShouldEqual, 2)

		columns, results, err = eql.Perform("LOOKUP .ID WITHIN .Shasum == {1}", "1234567890")
		SoMsg("shasum[1] lookup error", err, ShouldBeNil)
		SoMsg("shasum[1] results length", len(results), ShouldEqual, 1)
		SoMsg("shasum[1] results values", results, ShouldEqual, clContext.Contexts{
			{"id": int64(1)},
		})
		SoMsg("shasum[1] results columns", len(columns), ShouldEqual, 1)

		// placeholders
		columns, results, err = eql.Perform("LOOKUP .ID WITHIN .Shasum == {1}", "1234567890")
		SoMsg("shasum[2] lookup error", err, ShouldBeNil)
		SoMsg("shasum[2] results length", len(results), ShouldEqual, 1)
		SoMsg("shasum[2] results values", results, ShouldEqual, clContext.Contexts{
			{"id": int64(1)},
		})
		SoMsg("shasum[2] results columns", len(columns), ShouldEqual, 1)

		// datetime
		columns, results, err = eql.Perform("LOOKUP .Updated WITHIN .Shasum == {1}", "1234567890")
		SoMsg("shasum[3] lookup error", err, ShouldBeNil)
		SoMsg("shasum[3] results length", len(results), ShouldEqual, 1)
		SoMsg("shasum[3] results values", results, ShouldEqual, clContext.Contexts{
			{"updated": now123},
		})
		SoMsg("shasum[3] results columns", len(columns), ShouldEqual, 1)

		deletes := []struct {
			table    string
			err      Assertion
			id       int64
			affected int64
		}{
			{table: "page", err: ShouldBeNil, id: 1, affected: 1},
			{table: "page", err: ShouldBeNil, id: 2, affected: 1},
			{table: "nope", err: ShouldNotBeNil, id: 3, affected: 0},
			{table: "page", err: ShouldBeNil, id: 3, affected: 0},
			{table: "page", err: ShouldNotBeNil, id: 0, affected: 0},
		}

		tx, err = eql.SqlBegin()
		SoMsg("sql begin err", err, ShouldBeNil)
		SoMsg("sql transaction", tx, ShouldNotBeNil)

		for idx, test := range deletes {
			affected, err := tx.Delete(test.table, test.id)
			SoMsg(fmt.Sprintf("delete test #%d error", idx), err, test.err)
			SoMsg(fmt.Sprintf("delete test #%d affected", idx), affected, ShouldEqual, test.affected)
		}
		SoMsg("sql commit error", tx.Commit(), ShouldBeNil)

	})

	Convey("testdata/usecases", t, func() {

		batch := func(eql EnjinQL, dbh testdb.TestDB, basename, prefix string, a hrx.Archive) {
			if _, _, ok := a.Get("skip.test"); ok {
				return // skip these
			}

			var ok bool
			var inputEQL, outputSQL, outputERR string
			inputEQL, _, ok = a.Get("input.eql")
			SoMsg(prefix+"input.eql present", ok, ShouldBeTrue)
			inputEQL = strings.TrimSpace(inputEQL)
			SoMsg(prefix+"input.eql contents", inputEQL, ShouldNotEqual, "")

			if outputSQL, _, ok = a.Get("output.sql"); ok {
				outputSQL = strings.TrimSpace(outputSQL)
				outputSQL = strings.ReplaceAll(outputSQL, "\n", " ")
				SoMsg(prefix+"output.sql contents", outputSQL, ShouldNotEqual, "")
			} else if outputERR, _, ok = a.Get("output.err"); ok {
				outputERR = strings.TrimSpace(outputERR)
				SoMsg(prefix+"output.err contents", outputERR, ShouldNotEqual, "")
			}

			sqlQuery, _, err := eql.ToSQL(inputEQL)
			if outputSQL != "" {
				SoMsg(prefix+"output.sql error", err, ShouldBeNil)
				SoMsg(prefix+"output.sql equal", sqlQuery, ShouldEqual, outputSQL)
			} else if outputERR != "" {
				SoMsg(prefix+"output.err error", err, ShouldNotBeNil)
				SoMsg(prefix+"output.err equal", err.Error(), ShouldEqual, outputERR)
			}
		}

		for _, pathname := range td.LF("usecases") {
			basename := clPath.Base(pathname)
			prefix := basename + ": "
			a, err := hrx.ParseFile(pathname)
			SoMsg(prefix+"hrx parse file", err, ShouldBeNil)
			SoMsg(prefix+"hrx archive instance", a, ShouldNotBeNil)

			var eql EnjinQL
			var dbh testdb.TestDB
			if strings.HasPrefix(basename, "be-") {
				eql, dbh = makeBeEQL()
			} else if strings.HasPrefix(basename, "qf-") {
				eql, dbh = makeQfEQL()
			}
			SoMsg(prefix+"hrx filename is missing be- or qf prefix", eql, ShouldNotBeNil)
			SoMsg(prefix+"hrx filename is missing be- or qf prefix", dbh, ShouldNotBeNil)

			/*
				- input.eql: process an eql statement
				- output.sql: validate the results
				- output.err: expecting a specific error message
			*/

			Convey(basename, func() {
				if aa, ee := a.ParseHRX("batch.hrx"); ee == nil {

					for _, batchpath := range aa.List() {
						if strings.HasSuffix(batchpath, ".hrx") {
							batchname := clPath.Base(batchpath)
							batchprefix := basename + "[" + batchname + "]:"
							//Convey(batchname, func() {
							if ba, eee := aa.ParseHRX(batchpath); eee == nil {
								batch(eql, dbh, batchname, batchprefix, ba)
							}
							//})
						}
					}

				} else {
					batch(eql, dbh, basename, prefix, a)
				}
			})

			SoMsg(prefix+"close eql instance", eql.Close(), ShouldBeNil)
		}

	})
}

func makeBeEQL() (eql EnjinQL, dbh testdb.TestDB) {
	var err error
	tmpfile := tdata.TempFile("", "enjinql.*.be.db")
	if dbh, err = testdb.NewTestDB(); err != nil {
		panic(fmt.Errorf("%s: %v", tmpfile, err))
	}
	config := makeBeConfig()
	if eql, err = New(config, dbh.DBH(), dialects.Sqlite{}); err != nil {
		panic(fmt.Errorf("%s: %v", tmpfile, err))
	}
	makeBeData(eql)
	return
}

func makeBeData(eql EnjinQL) {
	tx, _ := eql.SqlBegin()
	defer tx.Commit()
	sid, _ := tx.Insert(PageSource, "1234567890", "en", "page", time.Now(), time.Now(), "/page-slug", "{}")
	permalink, _ := uuid.NewV4()
	short := shasum.MustBriefSum(permalink.String())
	_, _ = tx.Insert(PagePermalinkSource, sid, short, permalink.String())
	_, _ = tx.Insert(PageRedirectSource, sid, "/pg-slg")
	_, _ = tx.Insert("title", sid, "a page")
	_, _ = tx.Insert("description", sid, "this is a page")
	sid, _ = tx.Insert(PageSource, "0123456789", "en", "page", time.Now(), time.Now(), "/another-page", "{}")
	_, _ = tx.Insert("title", sid, "another page")
	_, _ = tx.Insert("description", sid, "this is another page")
}

func makeBeConfig() (config *Config) {
	config = NewConfig("be", "eql")
	for _, sc := range makeBeSourceConfigs() {
		config.AddSource(sc)
	}
	return
}

func makeBeSourceConfigs() (sources []*SourceConfig) {
	return []*SourceConfig{
		PageSourceConfig(),
		PagePermalinkSourceConfig(),
		PageRedirectSourceConfig(),
		{Name: "page_title", Parent: values.Ref(PageSource), Values: []*SourceConfigValue{
			{String: &SourceConfigValueString{
				Key:  "text",
				Size: 160,
			}},
		}},
		{Name: "page_description", Parent: values.Ref(PageSource), Values: []*SourceConfigValue{
			{String: &SourceConfigValueString{
				Key:  "text",
				Size: 160,
			}},
		}},
	}
}

func hasWord(eql EnjinQL, word string) (id int64) {
	if _, results, err := eql.Perform(`LOOKUP word.ID WITHIN word.Word == {1}`, word); err == nil && len(results) > 0 {
		if v, ok := results[0]["id"].(int64); ok {
			id = v
		}
	}
	return
}

func makeQfEQL() (eql EnjinQL, dbh testdb.TestDB) {
	var err error
	tmpfile := tdata.TempFile("", "enjinql.*.qf.db")
	if dbh, err = testdb.NewTestDB(); err != nil {
		panic(fmt.Errorf("%s: %v", tmpfile, err))
	}
	if eql, err = New(makeQfConfig(), dbh.DBH(), dialects.Sqlite{}); err != nil {
		panic(err)
	}
	makeBeData(eql)
	makeQfQuote(eql, "1122334455", "/1122334455aabbccddeeff", "this is the contents of a quote")
	makeQfQuote(eql, "0102000405", "/0102030405aabbccddeeff", "and this is the body of another quote")
	return
}

func makeQfQuote(eql EnjinQL, shasum, url, contents string) {
	tx, _ := eql.SqlBegin()
	defer tx.Commit()
	sid, _ := tx.Insert(PageSource, shasum, "en", "quote", url)
	_, _ = tx.Insert(PageRedirectSource, sid, "/"+shasum)
	fields := strings.Fields(contents)
	counts := make(map[string]int)
	for _, word := range fields {
		counts[word] += 1
	}
	unique := make(map[string]struct{})
	for _, word := range fields {
		var wid int64
		if wid = hasWord(eql, word); wid == 0 {
			wid, _ = tx.Insert("word", string(word[0]), word, word)
		}
		if _, present := unique[word]; present {
			continue
		}
		unique[word] = struct{}{}
		_, _ = tx.Insert("page_words", sid, wid, counts[word])
	}
}

func makeQfConfig() (config *Config) {
	config, _ = NewConfig("qf", "eql").
		AddSource(PageSourceConfig()).
		AddSource(PagePermalinkSourceConfig()).
		AddSource(PageRedirectSourceConfig()).
		// list of distinct lower-cased words
		NewSource("word").
		NewStringValue("letter", 1).
		NewStringValue("word", 256).
		NewStringValue("flat", 256).
		DoneSource().
		// associated words with specific pages
		NewSource("page_words").
		SetParent(PageSource).
		NewLinkedValue("word", SourceIdKey).
		NewIntValue("hits").
		DoneSource().
		NewSource("word_letters").
		SetParent("word").
		NewStringValue("letter", 1).
		DoneSource().
		Make()
	return
}
