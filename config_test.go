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
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/go-corelibs/hrx"
	clPath "github.com/go-corelibs/path"
	"github.com/go-corelibs/tdata"
	"github.com/go-corelibs/values"
)

func TestConfig(t *testing.T) {

	td := tdata.New()

	Convey("panic tests", t, func() {

		So(func() {
			_ = MakeSourceConfig("", "test", nil)
		}, ShouldPanic)

	})

	Convey("SourceConfigType", t, func() {

		So(SourceConfigType(254).String(), ShouldEqual, "unknown")
		So(UnknownSourceType.String(), ShouldEqual, "unknown")
		So(DataSourceType.String(), ShouldEqual, "data")
		So(LinkSourceType.String(), ShouldEqual, "link")
		So(JoinSourceType.String(), ShouldEqual, "join")

	})

	Convey("SourceConfigValue", t, func() {

		for idx, test := range []struct {
			input *SourceConfigValue
			name  string
			check string
		}{
			{NewIntValue("int"), "int", "int"},
			{NewBoolValue("bool"), "bool", "bool"},
			{NewTimeValue("time"), "time", "time"},
			{NewFloatValue("float"), "float", "float"},
			{NewStringValue("string", 1), "string", "string"},
			{NewLinkedValue("linked", "other"), "linked_other", "linked"},
			{&SourceConfigValue{}, "", "none"},
		} {

			prefix := fmt.Sprintf("test #%d ", idx)
			SoMsg(prefix+"(name)", test.input.Name(), ShouldEqual, test.name)
			if test.check == "none" {
				SoMsg(prefix+"(clone)", test.input.Clone(), ShouldBeNil)
			} else {
				SoMsg(prefix+"(clone)", test.input.Clone().Name(), ShouldEqual, test.name)
			}
			switch test.check {
			case "none":
				SoMsg(prefix+"("+test.check+")", test.input.Int, ShouldBeNil)
				SoMsg(prefix+"("+test.check+")", test.input.Bool, ShouldBeNil)
				SoMsg(prefix+"("+test.check+")", test.input.Time, ShouldBeNil)
				SoMsg(prefix+"("+test.check+")", test.input.Float, ShouldBeNil)
				SoMsg(prefix+"("+test.check+")", test.input.String, ShouldBeNil)
				SoMsg(prefix+"("+test.check+")", test.input.Linked, ShouldBeNil)
			case "int":
				SoMsg(prefix+"("+test.check+")", test.input.Int, ShouldNotBeNil)
			case "bool":
				SoMsg(prefix+"("+test.check+")", test.input.Bool, ShouldNotBeNil)
			case "time":
				SoMsg(prefix+"("+test.check+")", test.input.Time, ShouldNotBeNil)
			case "float":
				SoMsg(prefix+"("+test.check+")", test.input.Float, ShouldNotBeNil)
			case "string":
				SoMsg(prefix+"("+test.check+")", test.input.String, ShouldNotBeNil)
			case "linked":
				SoMsg(prefix+"("+test.check+")", test.input.Linked, ShouldNotBeNil)
			default:
				panic(fmt.Sprintf("qa developer error, invalid test.check value: %q", test.check))
			}

		}

	})

	Convey("SourceConfigValues", t, func() {

		csv := ConfigSourceValues{
			NewIntValue("int"),
			NewBoolValue("bool"),
		}
		So(csv.Names(), ShouldEqual, []string{"int", "bool"})
		So(csv.HasLinks(), ShouldBeFalse)

		csv = ConfigSourceValues{
			NewIntValue("int"),
			NewBoolValue("bool"),
			NewLinkedValue("linked", "other"),
		}
		So(csv.Names(), ShouldEqual, []string{"int", "bool", "linked_other"})
		So(csv.HasLinks(), ShouldBeTrue)

	})

	Convey("SourceConfig", t, func() {

		sc := &SourceConfig{Name: "test"}
		So(sc.Type(), ShouldEqual, DataSourceType)

		sc.Parent = values.Ref("page")
		So(sc.Type(), ShouldEqual, LinkSourceType)

		sc.Values = ConfigSourceValues{
			NewLinkedValue("page", "id"),
		}
		So(sc.Type(), ShouldEqual, JoinSourceType)

	})

	Convey("ConfigSources", t, func() {

		s := ConfigSources{
			&SourceConfig{Name: "empty", Unique: [][]string{}, Index: [][]string{}},
			&SourceConfig{Name: "linked", Parent: values.Ref("page")},
			&SourceConfig{Name: "joined", Parent: values.Ref("page"), Values: ConfigSourceValues{NewLinkedValue("page", "id")}},
		}

		So(s.Get("nope"), ShouldBeNil)
		So(s.Get("empty"), ShouldEqual, s[0])
		So(s.Names(), ShouldEqual, []string{"empty", "linked", "joined"})
		So(s.DataNames(), ShouldEqual, []string{"empty"})
		So(s.LinkNames(), ShouldEqual, []string{"linked"})
		So(s.JoinNames(), ShouldEqual, []string{"joined"})

	})

	Convey("testdata tests", t, func() {

		tests := td.LF("config")

		for _, pathname := range tests {
			prefix := clPath.Base(pathname) + ": "

			contents := td.F(pathname)
			SoMsg(prefix+"testdata contents", contents, ShouldNotEqual, "")

			a, err := hrx.ParseData(pathname, contents)
			SoMsg(prefix+"hrx parse error", err, ShouldBeNil)
			SoMsg(prefix+"hrx archive pointer", a, ShouldNotBeNil)

			configJson, _, ok := a.Get("config.json")
			SoMsg(prefix+"config.json present", ok, ShouldBeTrue)
			parsed, parseErr := ParseConfig(configJson)

			errorText, _, confirmErr := a.Get("error.txt")
			if confirmErr {

				SoMsg(prefix+"[ERR] parse error is nil", parseErr, ShouldNotBeNil)
				SoMsg(prefix+"[ERR] parse error message is incorrect", parseErr.Error(), ShouldEqual, errorText)
				SoMsg(prefix+"[ERR] parsed config is not nil", parsed, ShouldBeNil)

			} else {

				SoMsg(prefix+"[INF] parse error is not nil", parseErr, ShouldBeNil)
				SoMsg(prefix+"[INF] parsed config is nil", parsed, ShouldNotBeNil)
				SoMsg(prefix+"[INF] clone parsed config", parsed.Clone().String(), ShouldEqual, parsed.String())

				outputJson, _, checkOutputJson := a.Get("output.json")
				if checkOutputJson {
					SoMsg(prefix+"[INF] parsed string is incorrect", parsed.String(), ShouldEqual, outputJson)
				}

			}

		}

	})

	Convey("builder methods", t, func() {

		for idx, test := range []struct {
			label  string
			input  *Config
			output string
			err    Assertion
		}{

			{
				"new source, int value",
				NewConfig().NewSource("test").NewIntValue("present").DoneSource(),
				`{"sources":[{"name":"test","values":[{"int":{"key":"present"}}]}]}`,
				ShouldBeNil,
			},

			{
				"new source, bool value",
				NewConfig().NewSource("test").NewBoolValue("present").DoneSource(),
				`{"sources":[{"name":"test","values":[{"bool":{"key":"present"}}]}]}`,
				ShouldBeNil,
			},

			{
				"new source, float value",
				NewConfig().NewSource("test").NewFloatValue("present").DoneSource(),
				`{"sources":[{"name":"test","values":[{"float":{"key":"present"}}]}]}`,
				ShouldBeNil,
			},

			{
				"new source, string value",
				NewConfig().NewSource("test").NewStringValue("present", -1).DoneSource(),
				`{"sources":[{"name":"test","values":[{"string":{"key":"present","size":-1}}]}]}`,
				ShouldBeNil,
			},

			{
				"new source, linked value",
				NewConfig().
					NewSource("full").NewIntValue("counter").DoneSource().
					NewSource("test").NewLinkedValue("full", "presence").DoneSource(),
				`{"sources":[` +
					`{"name":"full","values":[{"int":{"key":"counter"}}]},` +
					`{"name":"test","values":[{"linked":{"table":"full","key":"presence"}}]}` +
					`]}`,
				ShouldBeNil,
			},

			{
				"new source, set parent",
				NewConfig().
					NewSource("full").NewIntValue("counter").DoneSource().
					NewSource("test").SetParent("full").NewIntValue("present").DoneSource(),
				`{"sources":[` +
					`{"name":"full","values":[{"int":{"key":"counter"}}]},` +
					`{"name":"test","parent":"full","values":[{"int":{"key":"present"}}]}` +
					`]}`,
				ShouldBeNil,
			},

			{
				"new source, add unique",
				NewConfig().
					NewSource("test").NewIntValue("present").AddUnique("present").DoneSource(),
				`{"sources":[{"name":"test","values":[{"int":{"key":"present"}}],"unique":[["present"]]}]}`,
				ShouldBeNil,
			},

			{
				"new source, add index",
				NewConfig().
					NewSource("test").NewIntValue("present").AddIndex("present").DoneSource(),
				`{"sources":[{"name":"test","values":[{"int":{"key":"present"}}],"index":[["present"]]}]}`,
				ShouldBeNil,
			},
		} {

			_, err := test.input.Make()
			SoMsg(
				fmt.Sprintf("#%d - %s", idx, test.label),
				err,
				test.err,
			)
			SoMsg(
				fmt.Sprintf("#%d - %s", idx, test.label),
				test.input.Serialize(),
				ShouldEqual,
				test.output,
			)

		}

	})
}
