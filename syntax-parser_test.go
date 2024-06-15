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
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-corelibs/hrx"
	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/tdata"
)

func oneLine(input string) (output string) {
	output = strings.ReplaceAll(input, "\n", " ")
	output = strings.TrimSpace(output)
	return
}

func TestSyntaxParser(t *testing.T) {
	td := tdata.New()

	Convey("bad structures", t, func() {

		// invalid operator structures
		op := &Operator{}
		SoMsg("op-nil string", op.String(), ShouldEqual, "")
		SoMsg("op-nil validate", op.validate(), ShouldNotBeNil)
		// *=
		cond, err := op.makeCS(nil, nil)
		SoMsg("op-nil makeCS err", err, ShouldNotBeNil)
		SoMsg("op-nil makeCS cond", cond, ShouldBeNil)
		// ~=
		cond, err = op.makeCF(nil, nil)
		SoMsg("op-nil makeCF.1 err", err, ShouldNotBeNil)
		SoMsg("op-nil makeCF.1 cond", cond, ShouldBeNil)
		cond, err = op.makeCF(nil, "")
		SoMsg("op-nil makeCF.2 err", err, ShouldNotBeNil)
		SoMsg("op-nil makeCF.2 cond", cond, ShouldBeNil)
		// ^=
		cond, err = op.makeSW(nil, nil)
		SoMsg("op-nil makeSW err", err, ShouldNotBeNil)
		SoMsg("op-nil makeSW cond", cond, ShouldBeNil)
		// $=
		cond, err = op.makeEW(nil, nil)
		SoMsg("op-nil makeEW err", err, ShouldNotBeNil)
		SoMsg("op-nil makeEW cond", cond, ShouldBeNil)
		// LIKE
		cond, err = op.makeLK(nil, nil)
		SoMsg("op-nil makeLK err", err, ShouldNotBeNil)
		SoMsg("op-nil makeLK cond", cond, ShouldBeNil)

	})

	Convey("testdata/syntax", t, func() {

		batch := func(prefix string, a hrx.Archive) {
			var ok bool
			var inputEQL, outputEQL, outputERR string
			inputEQL, _, ok = a.Get("input.eql")
			SoMsg(prefix+"input.eql ok", ok, ShouldBeTrue)
			inputEQL = oneLine(inputEQL)
			if outputEQL, _, ok = a.Get("output.eql"); ok {
				outputEQL = oneLine(outputEQL)
			} else {
				outputERR, _, ok = a.Get("output.err")
				SoMsg(prefix+"output.err and output.eql missing", ok, ShouldBeTrue)
				outputERR = oneLine(outputERR)
			}

			syntax, err := ParseSyntax(inputEQL)
			if outputERR == "" {
				SoMsg(prefix+" unexpected error", err, ShouldBeNil)
				SoMsg(prefix+" output.eql", syntax.String(), ShouldEqual, outputEQL)
				SoMsg(prefix+" syntax err", syntax.Validate(), ShouldBeNil)
			} else {
				SoMsg(prefix+" expected error", err, ShouldNotBeNil)
				// Skip for now: SoMsg(prefix+" output.err", err.Error(), ShouldEqual, outputERR)
			}
		}

		for _, pathname := range td.LF("syntax") {

			basename := clPath.Base(pathname)

			prefix := basename + ": "
			a, err := hrx.ParseFile(pathname)
			SoMsg(prefix+"hrx parse error", err, ShouldBeNil)
			SoMsg(prefix+"hrx archive", a, ShouldNotBeNil)

			/*
				- input.eql: text to parse
				- output.eql: expected .String
				- output.err: expected error message
			*/

			if aa, ee := a.ParseHRX("batch.hrx"); ee == nil {

				Convey(basename, func() {

					for _, batchpath := range aa.List() {

						if ba, eee := aa.ParseHRX(batchpath); eee == nil {
							batch(prefix, ba)
						}

					}

				})

			} else {
				batch(prefix, a)
			}

		}

	})

}
