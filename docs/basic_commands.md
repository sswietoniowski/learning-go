# Basic Commands

Useful commands for Go.

## General

To check a Go version, run:

```bash
go version
```

To get help for a command, run:

```bash
go help <command>
```

## Module Management

To create a Go module, run:

```bash
go mod init <module_name>
```

By common convention this module might be called `vendor` as it is used to store all the dependencies.

To clean up the Go module, run:

```bash
go mod tidy
```

To install a package, run:

```bash
go get <package_name>
```

## Build & Run

To run a Go program, run:

```bash
go run <file_name>
```

To build a Go program, run:

```bash
go build <file_name>
```

To install a Go program, run:

```bash
go install <file_name>
```

## Testing

Tests in Go are written in the same package as the code they test, but in a separate file with `_test` suffix.

To write a test, create a file with `_test` suffix and a function with `Test` prefix followed by the name of the function you want to test. As a parameter, this function takes a pointer to `testing.T` struct.

For example, if you want to test a function `Sum` in `sum.go` file:

```go
func Sum(a, b int) int {
    return a + b
}
```

You should create a file `sum_test.go` with a function `TestSum` that takes a pointer to `testing.T` struct as a parameter.

```go
func TestSum(t *testing.T) {
    got := Sum(1, 2)
    want := 3

    if got != want {
        t.Errorf("Sum(1, 2) = %v, expected = %v", got, want)
    }
}
```

To run tests, run:

```bash
go test
```

To run single test, run:

```bash
go test -run <test_name>
```

To run benchmarks, run:

```bash
go test -bench .
```

To run tests with coverage, run:

```bash
go test -cover
```

To run all tests in a directory, run:

```bash
go test ./...
```

## Documenting

To generate documentation, run:

```bash
go doc <package_name>
```

To generate documentation for a specific function, run:

```bash
go doc <package_name>.<function_name>
```

To generate documentation for a specific function in a specific file, run:

```bash
go doc <package_name>/<file_name>.<function_name>
```

Now, there is a question: _how to write documentation?_

The answer is simple: write comments.

```go
// Sum returns a sum of two integers.
func Sum(a, b int) int {
    return a + b
}
```

Above the function, write a comment that describes what the function does.

> The comment should start with the name of the function and then describe what it does. By convention you should add a period at the end of the comment. Generally we should write comments for all exported functions and types.

As a side note, we can also write comments for variables and constants (being part of the public API).

```go
// Pi is a mathematical constant.
const Pi = 3.14

// Name is a name of a person.
var Name = "John"
```

And for types (again public ones).

```go
// Person is a person.
type Person struct {
    Name string
    Age  int
}
```

Of course we can (and should) write comments for packages.

```go
// Package math provides basic mathematical functions.
package math
```

## Best Practices

To verify the code, run:

```bash
go vet
```

To check style, use `golint` (it's not the only one, but it's quite popular).

First, install it:

```bash
go install golang.org/x/lint/golint@latest
```

Then, run it:

```bash
golint
```

To format the code, run:

```bash
gofmt -w <file_name> # -w flag overwrites the file
gofmt -d <file_name> # -d flag shows the diff
```

Just to fix everything, run (it's the same as `gofmt -w -l .`):

```bash
go fmt .
```

Alternatively you can use `goimports` for formatting and fixing imports.

To do that first install it:

```bash
go install golang.org/x/tools/cmd/goimports@latest
```

Then, run it:

```bash
goimports -w -l .
```
