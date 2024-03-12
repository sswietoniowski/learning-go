# commander [![Build Status](https://travis-ci.com/shomali11/commander.svg?branch=master)](https://travis-ci.com/shomali11/commander) [![Go Report Card](https://goreportcard.com/badge/github.com/shomali11/commander)](https://goreportcard.com/report/github.com/shomali11/commander) [![GoDoc](https://godoc.org/github.com/shomali11/commander?status.svg)](https://godoc.org/github.com/shomali11/commander) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Command evaluator and parser

## Features

* Matches commands against provided text
* Extracts parameters from matching input
* Provides default values for missing parameters
* Supports String, Integer, Float and Boolean parameters
* Supports "word" {} vs "sentence" <> parameter matching

## Dependencies

* `proper` [github.com/shomali11/proper](https://github.com/shomali11/proper)


# Examples

## Example 1

In this example, we are matching a few strings against a command format, then parsing parameters if found or returning default values.

```go
package main

import (
	"fmt"
	"github.com/shomali11/commander"
)

func main() {
	properties, isMatch := commander.NewCommand("ping").Match("ping")
	fmt.Println(isMatch)            // true
	fmt.Println(properties)         // {}

	properties, isMatch = commander.NewCommand("ping").Match("pong")
	fmt.Println(isMatch)            // false
	fmt.Println(properties)         // nil

	properties, isMatch = commander.NewCommand("echo {word}").Match("echo hello world!")
	fmt.Println(isMatch)                               // true
	fmt.Println(properties.StringParam("word", ""))    // hello

	properties, isMatch = commander.NewCommand("echo <sentence>").Match("echo hello world!")
	fmt.Println(isMatch)                                   // true
	fmt.Println(properties.StringParam("sentence", ""))    // hello world!

	properties, isMatch = commander.NewCommand("repeat {word} {number}").Match("repeat hey 5")
	fmt.Println(isMatch)                                 // true
	fmt.Println(properties.StringParam("word", ""))      // hey
	fmt.Println(properties.IntegerParam("number", 0))    // 5

	properties, isMatch = commander.NewCommand("repeat {word} {number}").Match("repeat hey")
	fmt.Println(isMatch)                                 // true
	fmt.Println(properties.StringParam("word", ""))      // hey
	fmt.Println(properties.IntegerParam("number", 0))    // 0

	properties, isMatch = commander.NewCommand("search <stuff> {size}").Match("search hello there everyone 10")
	fmt.Println(isMatch)                                // true
	fmt.Println(properties.StringParam("stuff", ""))    // hello there everyone
	fmt.Println(properties.IntegerParam("size", 0))     // 10
}
```

## Example 2

In this example, we are tokenizing the command format and returning each token with a number that determines whether it is a parameter (word vs sentence) or not

```go
package main

import (
	"fmt"
	"github.com/shomali11/commander"
)

func main() {
	tokens := commander.NewCommand("echo {word} <sentence>").Tokenize()
	for _, token := range tokens {
		fmt.Println(token)
	}
}
```

Output:
```
&{echo NOT_PARAMETER}
&{word WORD_PARAMETER}
&{sentence SENTENCE_PARAMETER}
```