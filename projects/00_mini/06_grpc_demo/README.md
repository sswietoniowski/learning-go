# gRPC

This is a simple gRPC client-server application based on [this](https://grpc.io/docs/languages/go/quickstart/) and [this](https://grpc.io/docs/languages/go/basics/) introduction to gRPC.

- [gRPC](#grpc)
  - [Features](#features)
  - [Technologies](#technologies)
  - [Setup](#setup)

## Features

This application demonstrates the following features:

- simple RPC,
- server-side streaming RPC,
- client-side streaming RPC,
- bidirectional streaming RPC.

## Technologies

The application is built using the following technologies, libraries, frameworks, and tools:

- [Go](https://golang.org/),
- [protobuf-go](https://github.com/protocolbuffers/protobuf-go),
- [grpc-go](https://github.com/grpc/grpc-go).

## Setup

You need to install [Protobuf](https://protobuf.dev/) compiler to generate Go code from the `.proto` files. You can do this by running the following command:

```bash
sudo apt-get install -y protobuf-compiler
protoc --version  # Ensure compiler version is 3+
```

You might also install [Buf](https://github.com/bufbuild/buf) to lint the `.proto` files. You can do this by running the following command:

```bash
brew install bufbuild/buf/buf
```

Which is a linter for `.proto` files.

That can be done easily with help of [Homebrew](https://brew.sh/):

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

Before generating Go code from the `.proto` files, you need to install the following Go packages:

```bash
go get google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go
go get google.golang.org/grpc/cmd/protoc-gen-go-grpc
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
export PATH="$PATH:$(go env GOPATH)/bin"
```

To generate Go code from the `.proto` files, you can run the following command:

```bash
protoc \
  --go_out=./internal/common/genproto \
  --go_opt=module=github.com/sswietoniowski/learning-go/projects/00_mini/06_grpc_demo/internal/common/genproto \
  --go-grpc_out=./internal/common/genproto \
  --go-grpc_opt=module=github.com/sswietoniowski/learning-go/projects/00_mini/06_grpc_demo/internal/common/genproto \
  ./api/protobuf/*.proto
```

To run this application, you need to start the server first and then the client. You can do this by running the following commands:

To start the server:

```bash
go build -o ./build ./cmd/server && ./build/server
```

And then to start the client (separate terminal):

```bash
go build -o ./build ./cmd/client && ./build/client
```

If you want to add some extra dependencies to the project, you might need to run the following command (as we are using Go modules and vendoring) afterwards:

```bash
go mod tidy && go mod vendor
```
