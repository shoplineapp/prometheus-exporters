FROM golang:1.18.3 AS builder
WORKDIR /go/src

RUN apt-get update && apt-get install -y \
    curl \
    python3-pip \
    vim

RUN go install github.com/cosmtrek/air@latest
COPY go.mod go.sum ./
RUN go mod download
COPY ./cmd ./cmd
COPY ./interfaces ./interfaces
COPY ./internal ./internal
RUN go build -o exporter ./cmd/exporter.go

FROM golang:1.18.3
WORKDIR /go/src

RUN apt-get update && apt-get install -y python3-pip

RUN pip3 install 'mtools[mloginfo]'
RUN pip3 install python-dateutil
RUN mkdir -p tmp/logs

COPY --from=builder /go/src/exporter ./
CMD ["./exporter"]
