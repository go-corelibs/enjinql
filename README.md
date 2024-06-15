[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/go-corelibs/enjinql)
[![codecov](https://codecov.io/gh/go-corelibs/enjinql/graph/badge.svg?token=TudEftLKUz)](https://codecov.io/gh/go-corelibs/enjinql)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-corelibs/enjinql)](https://goreportcard.com/report/github.com/go-corelibs/enjinql)

# enjinql - Enjin Query Language

enjinql is the reference implementation of the Enjin Query Language, a feature
of Go-Enjin that manages the indexing and accessing of page content.

# Notice

While this project is in active use within the Go-Enjin project and example
sites, enjinql does not have a sufficient amount of unit tests and the syntax
has not been ratified as a formal specification.

Please consider enjinql as a proof-of-concept that is already on its way to
being a minimum viable product and for the time being, probably not the best
choice to use in non-Go-Enjin projects.

# Installation

``` shell
> go get github.com/go-corelibs/enjinql@latest
```

## Command Line Interface

``` shell
> go install github.com/go-corelibs/enjinql/cmd/enjinql@latest
```

# Go-CoreLibs

[Go-CoreLibs] is a repository of shared code between the [Go-Curses] and
[Go-Enjin] projects.

# License

```
Copyright 2024 The Go-CoreLibs Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use file except in compliance with the License.
You may obtain a copy of the license at

 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

[Go-CoreLibs]: https://github.com/go-corelibs
[Go-Curses]: https://github.com/go-curses
[Go-Enjin]: https://github.com/go-enjin
