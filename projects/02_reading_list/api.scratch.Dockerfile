ARG GO_VERSION=1.22.0

# STAGE 1: building the executable
FROM golang:${GO_VERSION}-alpine AS build

RUN apk add --no-cache git
RUN apk --no-cache add ca-certificates

# add a user here because addgroup and adduser are not available in scratch
RUN addgroup -S api && adduser -S -u 10000 -g api api

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

# build the executable
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /app/api ./cmd/api

# STAGE 2: build the container to run
FROM scratch AS final

LABEL maintainer="sswietoniowski"

# copy certs and passwd file
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd

# copy compiled app
COPY --from=build --chown=api:api /app/api /api

# environment variables
ENV DB_HOST=localhost
ENV DB_PORT=5432
ENV DB_USER=postgres
ENV DB_PASSWORD=PUT_REAL_PASSWORD_HERE
ENV DB_NAME=readinglist

# use the user we created in the first stage
USER api

ENTRYPOINT ["/api", "--port", "4000", "--env", "development", "--db", "postgresql", "--frontend", "http://web:8080"]
