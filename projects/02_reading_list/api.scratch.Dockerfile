ARG GO_VERSION=1.22.0

# STAGE 1: building the executable
FROM golang:${GO_VERSION}-alpine AS build

# install git, ca-certificates, create user & group
# one commnad to reduce the number of layers
RUN apk add --no-cache git \
    && apk --no-cache add ca-certificates \
    && addgroup -S api && adduser -S -u 10000 -g api api

# create a directory for the app
WORKDIR /src

# copy go.mod and go.sum files and download dependencies
COPY ./go.mod ./go.sum ./
RUN go mod download

# copy the rest of the files
COPY ./ ./

# build the executable (CGO_ENABLED=0 to build a static binary)
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /app/api ./cmd/api

# STAGE 2: build the container to run
FROM scratch AS final

# define metadata
LABEL maintainer="sswietoniowski"

# copy certs and passwd file from the build stage
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd

# copy compiled app from the build stage
COPY --from=build --chown=api:api /app/api /api

# define environment variables
ENV DB_HOST=localhost
ENV DB_PORT=5432
ENV DB_USER=postgres
ENV DB_PASSWORD=PUT_REAL_PASSWORD_HERE
ENV DB_NAME=readinglist

# use the user we created in the first stage
USER api

# run the app using ENTRYPOINT and CMD (default flags that can be overwritten) combination
ENTRYPOINT ["/api"]
CMD ["--port", "4000", "--env", "development", "--db", "postgresql", "--frontend", "http://web:8080"]
