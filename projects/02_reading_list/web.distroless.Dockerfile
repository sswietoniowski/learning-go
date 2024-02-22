ARG GO_VERSION=1.22.0

# STAGE 1: building the executable
FROM golang:${GO_VERSION}-alpine AS build

# install git, create user & group
# one commnad to reduce the number of layers
RUN apk add --no-cache git \
    && addgroup -S web && adduser -S web -G web

# create a directory for the app
WORKDIR /src

# copy go.mod and go.sum files and download dependencies
COPY ./go.mod ./go.sum ./
RUN go mod download

# copy the rest of the files
COPY ./ ./

# Build the executable (CGO_ENABLED=0 to build a static binary)
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /app/web ./cmd/web

# STAGE 2: build the container to run
FROM gcr.io/distroless/static AS final

# define metadata
LABEL maintainer="sswietoniowski"

# copy passwd file from the build stage
COPY --from=build /etc/passwd /etc/passwd

# copy compiled app from the build stage
COPY --from=build --chown=web:web /app/web /web
COPY --from=build --chown=web:web /src/ui /ui

# use the user we created in the first stage
USER web

# run the app using ENTRYPOINT and CMD (default flags that can be overwritten) combination
ENTRYPOINT ["/web"]
CMD ["--port", "8080", "--env", "development", "--backend", "http://api:4000/api/v1"]

