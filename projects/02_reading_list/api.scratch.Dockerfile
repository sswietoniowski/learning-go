ARG GO_VERSION=1.21.6

# STAGE 1: building the executable
FROM golang:${GO_VERSION}-alpine AS build
RUN apk add --no-cache git
RUN apk --no-cache add ca-certificates

# add a user here because addgroup and adduser are not available in scratch
RUN addgroup -S api \
    && adduser -S -u 10000 -g api api

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

# build the executable
RUN CGO_ENABLED=0 go build \
    -installsuffix 'static' \
    -o /api ./cmd/api

# STAGE 2: build the container to run
FROM scratch AS final

LABEL maintainer="sswietoniowski"

# copy compiled app
COPY --from=build /api /api

# copy certs and passwd file
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd

ENV APP_PORT=4000
ENV APP_ENV=production
ENV APP_DB=postgresql
ENV APP_FRONTEND=http://localhost:8080
ENV DB_HOST=localhost
ENV DB_PORT=5432
ENV DB_USER=postgres
ENV DB_PASSWORD=PUT_REAL_PASSWORD_HERE
ENV DB_NAME=readinglist

USER api

# TODO: fix the entrypoint
#ENTRYPOINT ["/api", "--port", "$APP_PORT", "--env", "$APP_ENV", "--db", "$APP_DB", "--frontend", "$APP_FRONTEND"]
ENTRYPOINT ["/api", "--port", "4000", "--env", "development", "--db", "postgresql", "--frontend", "http://web:8080"]