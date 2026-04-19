FROM golang:1.26-alpine3.22

WORKDIR /src/backend

RUN go install github.com/cespare/reflex@v0.3.1

CMD ["reflex", "-c", "reflex.conf"]
