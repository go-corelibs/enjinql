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

// Package enjinql provides an abstracted system for indexing and querying a
// collection of content.
//
// Sources is an interface for managing and querying a collection of Source
// instances.
//
// Source is an interface for managing the correlation of contextual
// information with the collection of content.
//
// # Source Forms
//
// There are currently three forms of what a Source is, conceptually speaking.
//
// The first is the "primary source" that the Enjin Query Language is tasked
// with indexing.
// This is not the actual data being stored in a table per-se, just the
// necessary bits for providing a reference back to the actual content (like
// most other indexing systems).
// For example, in the use of the fantastic [bleve search] system, the
// site developer must provide a mapping of fields to index and then for each
// url of content, pass the url and its specific fields to bleve to index.
// bleve analyses the individual fields of content and using its internal
// primary source (likely a btree or some other wizardry), links this listing
// of URLs to the content fields provided.
// Once the site indexing process is complete, the developer can easily produce
// a search page which end-users can use the bleve search syntax to find content
// present in the site.
//
// For the Enjin Query Language, more than just the URL is important.
// There is still the URL of course but there is a greater context present as
// well.
// There is the page's shasum identifier, the page's type of content, language
// and so on to form a contextual index for finding content.
// This isn't to replace bleve within Go-Enjin; end-user search features are
// notoriously difficult to get right and bleve excels at getting this right.
// As a developer of a Go-Enjin based system, writing features often requires
// querying the enjin for a contextual case that may not involve the URL at
// all and to be forced to query for a URL in order to query for a page's
// specific context becomes a rather tedious repetition of code that isn't
// easily abstracted.
// Cases like that are examples of what could be called the farming of technical
// debt.
// Of course developers don't like farming technical debt and because one of
// the main purposes of the Go-Enjin project is to cater to the needs of the
// developer rather than just the employers and end-users, Go-Enjin
// necessitates a system for managing and querying a context of information.
//
// The second form of Source is the "data source". These are sources of data
// that relate to the primary source but do not actually have any direct
// correlation with the primary source context.
// For example, if the developer needs to build up a list of distinct words
// found within any given page of content, it is beneficial to have an index
// of all distinct words in general and to then use another form of source
// which joins the data source with the primary source.
//
// The final form of Source is the "link source" and as noted in the data
// source description, these sources typically have only unique identifiers
// which link two data sources together.
// This is the basis for how the Enjin Query Language can construct SQL
// queries with multiple JOIN statements without the developer having to spell
// out the table relationships in their use of the Enjin Query Language.
//
// # Real World Example
//
// One of the Go-Enjin website projects is an experiment in exploring human
// thought through the strange indexing of a large set of quotes from various
// sources.
// This website is called [Quoted.FYI] and in specific, it allows visitors to
// select from a list of "first words" from all quotes and then select from
// another list of "next words", and so on, building up to a narrow selection
// of quotes sharing the same series of opening words.
// To implement this feature, systems like bleve just aren't designed for it.
// Let's take a little walkthrough of the Quoted.FYI website project.
//
// ## Quotes
//
//   - a quote is just a normal Go-Enjin page with a custom type of "quote"
//   - each quote has a context consisting of an author, a list of topics, the
//     body of the quote and a SHA1 hashing of the content (used to help avoid
//     exact duplicates though "plural specificity" is allowed)
//
// ## Statistics
//
//   - 482,673 quotes
//   - 23,731 authors
//   - 11,980 topics
//   - 98,879 words
//
// ## EnjinQL Setup
//
// Quoted.FYI of course uses the default Go-Enjin primary source which
// consists of a number of specific columns of information related to any
// given page.
//
//	| (Page Primary Source)                                  |
//	| Key      | Description                                 |
//	----------------------------------------------------------
//	| id       | an integer primary key                      |
//	| shasum   | specific version of a specific page         |
//	| language | language code of a specific page            |
//	| url      | relative URL path to a specific page        |
//	| stub     | JSON blob used to construct a specific page |
//
// These are the commonly used context keys when implementing Go-Enjin
// features, and so they're all a part of the primary enjinql source.
//
// For each of the authors, topics and words information, Quoted.FYI needs
// additional indexing to support these various things.
// They're relatively the same but let's take a look at the words indexing in
// more detail.
//
//	| (Word Data Source)                                          |
//	| Key     | Description                                       |
//	---------------------------------------------------------------
//	| id      | an integer primary key                            |
//	| word    | lower-cased plain text, unique, word              |
//	| flat    | a snake_cased version of word used in legacy URLs |
//
// The above data source, describes the list of distinct words used to
// consolidate the space required by the underlying database.
//
// So, now there are two sources defined, the primary one and the list of
// unique words. These two sources, being entirely unrelated to each other
// would result in a query error if used together.
//
// For example, to look up a list of page shasums for the "en" language:
//
//	(EQL) LOOKUP .Shasum WITHIN .Language == "en"
//	(SQL) SELECT be_eql_page.shasum
//	      FROM be_eql_page
//	      WHERE be_eql_page.language = "en";
//
// The above query would work and return the list of all page shasums that
// are of the "en" language.
//
//	(EQL) LOOKUP word.Word WITHIN .Word ^= "a"
//	(SQL) SELECT be_eql_word.word
//	      FROM be_eql_word
//	      WHERE be_eql_word LIKE "a%";
//
// The above query again would work, this time returning a list of all words
// starting with the letter "a". While not the most efficient due to the use
// of the "starts with" (^=) operator, it does perform well.
//
// Given just the primary and word data sources, the following query would
// not work (yet):
//
//	(EQL) LOOKUP .Shasum WITHIN word.Word == "thing"
//	(SQL) error: "be_eql_page" is not linked with "be_eql_word"
//
// The above query is supposed to return a list of page shasums where the page
// uses the word "thing" at least once.
// To make this work, the needs to be a link source connecting the
// relationship of pages to words within each page.
//
//	| (Page Words Link Source)                                  |
//	| Key     | Description                                     |
//	-------------------------------------------------------------
//	| id      | an integer primary key                          |
//	| page_id | id of the primary source                        |
//	| word_id | id of the word data source                      |
//	| hits    | how many times the word is used within the page |
//
// The above link source joins the primary source with the word data source
// and includes an extra count of how many times that specific word is used
// within the specific quote.
//
// Now, the query to get the list of pages with the word "thing":
//
//	(EQL) LOOKUP .Shasum WITHIN word.Word == "thing"
//	(SQL) SELECT be_eql_page.shasum
//	      FROM be_eql_page
//	      INNER JOIN be_eql_page_words ON be_eql_page.id=be_eql_page_words.page_id
//	      INNER JOIN be_eql_words      ON be_eql_word.id=be_eql_page_words.word_id
//	      WHERE be_eql_word.word == "thing";
//
// This demonstrates the simplicity of the Enjin Query Language in that the
// EQL statements don't need to do an SQL magic directly, such as sorting out
// the table joins.
// This is all done by the developer simply defining the various sources and
// then populating them with the content available.
//
// The below code demonstrates how to create the primary and word data sources
// depicted above:
//
//	// this is the top-level interface for interacting with the enjinql module
//	sources := enjinql.NewSources("be_eql")
//	// build the primary source
//	bePage, err := sources.newSource("page").
//	  StringValue("shasum", 10).               // shasum primary value (10 bytes)
//	  AddStringValue("language", 7).           // page language code (7 bytes)
//	  AddStringValue("type", 32).              // custom page type (32 bytes)
//	  AddStringValue("url", 2000).             // page url (2000 bytes)
//	  AddStringValue("stub", -1).              // enjin stub (default TEXT size)
//	  AddUnique("shasum", "language", "url").  // add a UNIQUE constraint on shasum + language + url
//	  AddIndex("shasum").                      // add a specific shasum sql index
//	  AddIndex("language").                    // add a specific language sql index
//	  AddIndex("type").                        // add a specific type sql index
//	  AddIndex("url").                         // add a specific url sql index
//	  Make()                                   // make the source instance
//	// build the word data source
//	beWord, err := sources.newSource("word").
//	  StringValue("word", 256).                // word primary value (256 bytes)
//	  AddStringValue("flat", 256).             // flat word value (256 bytes)
//	  AddUnique("word").                       // add a UNIQUE constraint on word
//	  AddIndex("word").                        // add a specific word sql index
//	  AddIndex("flat").                        // add a specific flat sql index
//	  Make()                                   // make the source instance
//	// build the page-word link source
//	bePageWords, err := bePage.newSource("words").
//	  LinkedValue("page", "id").               // page_id link to bePage source, id column
//	  AddLinkedValue("word", "id").            // word_id link to "word" source, id column
//	  AddIntValue("hits").                     // additional integer value
//	  AddUnique("page_id", "word_id").         // add a UNIQUE constraint on page_id + word_id
//	  AddIndex("page_id").                     // add a specific page_id index
//	  AddIndex("word_id").                     // add a specific word_id index
//	  Make()                                   // make the source instance
//
// With the above constructed, the developer can now proceed with updating the
// sources instance with the content available.
//
//	// add a new page to the primary source
//	sid, err = bePage.Insert("1234567890", "en", "quote", "/q/a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2", "{...json...}")
//	// for each of the distinct words present in the quote body, count the number of times the word is used, flatten
//	// the word and add it all to the word source
//	wid, err = beWord.InsertOrIgnore("word", "i'm", "i_m")
//	_, err = bePageWord.Insert(sid, wid, count)
//
// That process needs to be expanded up of course for the complete site, but
// for just the primary, data and link sources defined so far, the enjinql
// instance is now ready to build the actual SQL from parsing EQL statements:
//
//	// get a list of shasums for all pages with the word "thing"
//	sql, argv, err = sources.Parse(`LOOKUP .Shasum WITHIN word.Word == "thing"`)
//	// sql  => "SELECT ... WHERE be_eql_word.word=?;"
//	// argv => []interface{"thing"}
//	// err  => nil
//
// [bleve search]: https://github.com/blevesearch/bleve
// [Quoted.FYI]: https://github.com/go-enjin/website-quoted-fyi
package enjinql
