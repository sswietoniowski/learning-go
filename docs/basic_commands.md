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

In Go a module is a collection of Go packages stored in a file tree with a `go.mod` file at its root. The `go.mod` file defines the module's module path, which is also the import path used for the root directory, and its dependency requirements, which are the other modules needed for a successful build.

Example of a `go.mod` file:

```go
module github.com/username/repo

go 1.21

require (
    github.com/some/dependency v1.2.3
)
```

We should not edit the `go.mod` file manually. Instead, we should use the `go get` command to add a new dependency to the module.

If we're working on the dependency itself, we should use the `replace` directive in the `go.mod` file to point to the local copy of the dependency.

```go
module github.com/username/repo

go 1.21

require (
    github.com/some/dependency v1.2.3
)

replace github.com/some/dependency => ../dependency
```

To clean up the Go module, run:

```bash
go mod tidy
```

To install a package (add a dependency to the module), run:

```bash
go get <package_name>
```

To install specific version of a package, run:

```bash
go get <package_name>@<version>
```

To upgrade all dependencies (to the latest version), run:

```bash
go get -u
```

## Build & Run

To run a Go program, run:

```bash
go run <file_name>
```

To build a Go program (for the current module), run:

```bash
go build
```

By default the output file is named after the directory in which the `go build` command is run. To change the name of the output file, run:

```bash
go build -o <output_file_name>
```

We might build our program and then run it. To do that, run:

```bash
go build -o <output_file_name> && ./<output_file_name>
```

To install a Go program, run:

```bash
go install
```

After installing, the program is available for execution from any directory.

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

## Versioning

While working with Go, we should use semantic versioning.

The version number should be in the form of `vMAJOR.MINOR.PATCH` (e.g. `v1.2.3`).

[Modules versions](https://go.dev/doc/modules/version-numbers) 0 to 1 are using the following scheme:

- `v0.x.x` - initial development,
- `v1.x.x` - stable,
- `v1.5.x` - minor changes,
- `v1.5.1` - patch changes,
- `v1.6.0-beta.1` - pre-release version.

[Modules versions 2](https://go.dev/wiki/Modules#releasing-modules-v2-or-higher) and higher are using the following scheme:

- `module/v2 v2.x.x` - a new version with breaking changes.
- `module/v2 v2.0.0-beta.1` - pre-release version.
- `module/v2 v2.0.0-rc.1` - release candidate.
- `module/v2 v2.0.0` - stable.

Modules can be versioned using tags.

To create a tag, run:

```bash
git tag v1.2.3
```

[Here](https://blog.jetbrains.com/go/2020/03/25/working-with-go-modules-versioning/) we find great explanation how to work with Go modules and versioning in GoLand.

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
