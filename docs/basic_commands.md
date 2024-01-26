# Basic Commands

Useful commands for Go.

## General

To check a Go version, run:

```bash
go version
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

To run benchmarks, run:

```bash
go test -bench .
```

To run tests with coverage, run:

```bash
go test -cover
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
