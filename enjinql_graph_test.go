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
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSourceGraph(t *testing.T) {
	Convey("basics", t, func() {

		sg := newSourceGraph()

		err := sg.Add(
			newSourceNodeData("page"),
			&gSourceNode{
				name: "permalink",
				// INNER JOIN permalink ON page.id=permalink.page_id
				parent: &gSourceJoin{
					table: "permalink",
					this:  newSourceTableKey("permalink", "page_id"),
					other: newSourceTableKey("page", "id"),
				},
				link: make(map[string]*gSourceJoin),
			},
			&gSourceNode{
				name: "redirect",
				parent: &gSourceJoin{
					table: "redirect",
					this:  newSourceTableKey("redirect", "page_id"),
					other: newSourceTableKey("page", "id"),
				},
				link: make(map[string]*gSourceJoin),
			},
			&gSourceNode{
				name: "title",
				parent: newSourceJoin(
					"title", "page_id",
					newSourceTableKey("page", "id"),
				),
				link: make(map[string]*gSourceJoin),
			},
			newSourceNodeData("word"),
			&gSourceNode{
				name: "page_words",
				parent: newSourceJoin(
					"page_words", "page_id",
					newSourceTableKey("page", "id"),
				),
				link: map[string]*gSourceJoin{
					"word": newSourceJoin(
						"word", "id",
						newSourceTableKey("page_words", "word_id"),
					),
				},
			},
			&gSourceNode{
				name: "word_letters",
				parent: newSourceJoin(
					"word_letters", "word_id",
					newSourceTableKey("word", "id"),
				),
				link: map[string]*gSourceJoin{},
			},
			/* broken link with other link */
			//&gSourceNode{
			//	name:   "broken",
			//	parent: nil,
			//	link: map[string]*gSourceJoin{
			//		"other": newSourceJoin("broken", "other_id", newSourceTableKey("other", "id")),
			//	},
			//},
			//&gSourceNode{
			//	name:   "other",
			//	parent: nil,
			//	link: map[string]*gSourceJoin{
			//		"broken": newSourceJoin("other", "broken_id", newSourceTableKey("broken", "id")),
			//	},
			//},
			/* broken parent with other parent */
			//&gSourceNode{
			//	name:   "broken",
			//	parent: newSourceJoin("broken", "other_id", newSourceTableKey("other", "id")),
			//	link:   map[string]*gSourceJoin{},
			//},
			//&gSourceNode{
			//	name:   "other",
			//	parent: newSourceJoin("other", "broken_id", newSourceTableKey("broken", "id")),
			//	link:   map[string]*gSourceJoin{},
			//},
			/* broken parent with other link */
			//&gSourceNode{
			//	name:   "broken",
			//	parent: newSourceJoin("broken", "other_id", newSourceTableKey("other", "id")),
			//	link:   map[string]*gSourceJoin{},
			//},
			//&gSourceNode{
			//	name:   "other",
			//	parent: nil,
			//	link: map[string]*gSourceJoin{
			//		"broken": newSourceJoin("other", "broken_id", newSourceTableKey("broken", "id")),
			//	},
			//},
		)
		SoMsg("source graph add: err", err, ShouldBeNil)
		SoMsg("source graph validate: err", sg.validate(), ShouldBeNil)

		for _, test := range []struct {
			label   string
			sources []string
			err     Assertion
			plan    *gSourcePlan
		}{
			{
				label:   "only one link",
				sources: []string{"title"},
				err:     ShouldBeNil,
				plan:    &gSourcePlan{top: "title"},
			},

			{
				label:   "many links",
				sources: []string{"title", "permalink", "redirect"},
				err:     ShouldBeNil,
				plan: &gSourcePlan{
					top: "page",
					joins: []*gSourceJoin{
						{
							table: "title",
							this:  newSourceTableKey("title", "page_id"),
							other: newSourceTableKey("page", "id"),
						},
						{
							table: "permalink",
							this:  newSourceTableKey("permalink", "page_id"),
							other: newSourceTableKey("page", "id"),
						},
						{
							table: "redirect",
							this:  newSourceTableKey("redirect", "page_id"),
							other: newSourceTableKey("page", "id"),
						},
					},
				},
			},

			{
				label:   "one data, one link",
				sources: []string{"word", "title"},
				err:     ShouldBeNil,
				plan: &gSourcePlan{
					top: "page",
					joins: []*gSourceJoin{
						{
							table: "page_words",
							this:  newSourceTableKey("page_words", "page_id"),
							other: newSourceTableKey("page", "id"),
						},
						{
							table: "word",
							this:  newSourceTableKey("word", "id"),
							other: newSourceTableKey("page_words", "word_id"),
						},
						{
							table: "title",
							this:  newSourceTableKey("title", "page_id"),
							other: newSourceTableKey("page", "id"),
						},
					},
				},
			},

			{
				label:   "two data",
				sources: []string{"word", "page"},
				err:     ShouldBeNil,
				plan: &gSourcePlan{
					top: "page",
					joins: []*gSourceJoin{
						{
							table: "page_words",
							this:  newSourceTableKey("page_words", "page_id"),
							other: newSourceTableKey("page", "id"),
						},
						{
							table: "word",
							this:  newSourceTableKey("word", "id"),
							other: newSourceTableKey("page_words", "word_id"),
						},
					},
				},
			},

			{
				label:   "two non-primary sources",
				sources: []string{"word", "word_letters"},
				err:     ShouldBeNil,
				plan: &gSourcePlan{
					top: "word",
					joins: []*gSourceJoin{
						{
							table: "word_letters",
							this:  newSourceTableKey("word_letters", "word_id"),
							other: newSourceTableKey("word", "id"),
						},
					},
				},
			},

			{
				label:   "two non-primary, one pseudo-primary sources",
				sources: []string{"word", "word_letters", "permalink"},
				err:     ShouldBeNil,
				plan: &gSourcePlan{
					top: "page",
					joins: []*gSourceJoin{
						{
							table: "page_words",
							this:  newSourceTableKey("page_words", "page_id"),
							other: newSourceTableKey("page", "id"),
						},
						{
							table: "word",
							this:  newSourceTableKey("word", "id"),
							other: newSourceTableKey("page_words", "word_id"),
						},
						{
							table: "word_letters",
							this:  newSourceTableKey("word_letters", "word_id"),
							other: newSourceTableKey("word", "id"),
						},
						{
							table: "permalink",
							this:  newSourceTableKey("permalink", "page_id"),
							other: newSourceTableKey("page", "id"),
						},
					},
				},
			},

			{
				label:   "qf pages with word",
				sources: []string{"page", "word"},
				err:     ShouldBeNil,
				plan: &gSourcePlan{
					top: "page",
					joins: []*gSourceJoin{
						{
							table: "page_words",
							this:  newSourceTableKey("page_words", "page_id"),
							other: newSourceTableKey("page", "id"),
						},
						{
							table: "word",
							this:  newSourceTableKey("word", "id"),
							other: newSourceTableKey("page_words", "word_id"),
						},
					},
				},
			},

			{
				label:   "qf pages with words that start with letter",
				sources: []string{"page", "word_letters"},
				err:     ShouldBeNil,
				plan: &gSourcePlan{
					top: "page",
					joins: []*gSourceJoin{
						{
							table: "page_words",
							this:  newSourceTableKey("page_words", "page_id"),
							other: newSourceTableKey("page", "id"),
						},
						{
							table: "word",
							this:  newSourceTableKey("word", "id"),
							other: newSourceTableKey("page_words", "word_id"),
						},
						{
							table: "word_letters",
							this:  newSourceTableKey("word_letters", "word_id"),
							other: newSourceTableKey("word", "id"),
						},
					},
				},
			},
		} {

			plan, err := sg.plan(test.sources...)
			SoMsg("plan "+test.label+": err", err, test.err)
			SoMsg("plan "+test.label+": top", plan.top, ShouldEqual, test.plan.top)
			SoMsg("plan "+test.label+": yes", plan.String(), ShouldEqual, test.plan.String())
		}

	})
}
