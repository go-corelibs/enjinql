#!/usr/bin/make --no-print-directory --jobs=1 --environment-overrides -f

CORELIB_PKG := go-corelibs/enjinql
VERSION_TAGS += MAIN
MAIN_MK_SUMMARY := ${CORELIB_PKG}
MAIN_MK_VERSION := v0.0.0

BUILD_COMMANDS += enjinql
#BUILD_TAGS     += slower_lexer

GOTESTS_TAGS += testdb
CONVEY_TIMEOUT := 10s

#: installing participle requires cloning the repo and manually building from
#: within that source tree
#
### $ go install github.com/alecthomas/participle/cmd/participle@latest
### go: github.com/alecthomas/participle/cmd/participle@latest (in github.com/alecthomas/participle/cmd/participle@v0.0.0-20240509095130-5f96b0729ffe):
### 	The go.mod file for the module providing named packages contains one or
### 	more replace directives. It must not contain directives that would cause
### 	it to be interpreted differently than if it were the main module
#
#DEPS += github.com/alecthomas/participle/cmd/participle

include CoreLibs.mk

# if the railroad command is present in the PATH, include the "ebnf" target
# which produces a railroad diagram of the enjinql syntax
ifneq (,$(wildcard $(shell which railroad)))
ebnf: build
	@echo "# generating EBNF railroad diagram"
	@rm -rf _railroad || true
	@mkdir -vp _railroad
	@pushd _railroad > /dev/null; \
		../enjinql ebnf | railroad -w -o ./index.html; \
		perl -pi \
			-e "s,(href|src)='(railroad-diagrams\.(?:css|js))',\$$1='./\$$2',g" \
			./index.html \
			|| true; \
	popd > /dev/null
endif

ifneq (,$(wildcard $(shell which participle)))
lexer: build
	@echo "# generating faster participle lexer"
	@[ -f syntax-parser_faster.go.nope ] \
		&& mv -v syntax-parser_faster.go syntax-parser_faster.go.nope \
		|| true
	@./enjinql lexer \
		| participle gen lexer enjinql --name gGenerated \
		| gofmt > syntax-parser_faster_lexer.go
endif
