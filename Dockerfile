FROM golang:alpine AS builder

WORKDIR /go/src/current
COPY go.mod /go/src/current

RUN go mod download

COPY . /go/src/current

WORKDIR /go/src/current
RUN go build -v

FROM alpine:latest

COPY --from=builder /go/src/current/lightsites /lightsites

CMD ["/lightsites"]
