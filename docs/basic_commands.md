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

To run tests, run:

```bash
go test
```

To run benchmarks, run:

```bash
go test -bench .
```
