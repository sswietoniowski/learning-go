ARG GO_VERSION=1.21.6

# STAGE 1: building the executable
FROM golang:${GO_VERSION}-alpine AS build

RUN apk add --no-cache git

WORKDIR /src

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

# Build the executable
RUN CGO_ENABLED=0 go build \
    -installsuffix 'static' \
    -o /web ./cmd/web

# STAGE 2: build the container to run
FROM gcr.io/distroless/static AS final

LABEL maintainer="sswietoniowski"

# copy compiled app
COPY --from=build --chown=web:web /web /web
COPY --from=build --chown=web:web /src/ui /ui

ENV APP_PORT=4000
ENV APP_ENV=production
ENV APP_BACKEND=http://localhost:4000/api/v1

# TODO: fix the user
#RUN addgroup -S web && adduser -S web -G web
#USER web:web

# TODO: fix the entrypoint
#ENTRYPOINT ["/web", "--port", "$APP_PORT", "--env", "$APP_ENV", "--backend", "$APP_BACKEND"]
ENTRYPOINT ["/web", "--port", "8080", "--env", "development", "--backend", "http://api:4000/api/v1"]

