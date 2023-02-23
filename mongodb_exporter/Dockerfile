FROM golang:1.20.1 AS builder
WORKDIR /go/src

RUN apt-get update && apt-get install -y \
    git \
    curl \
    python3-pip \
    vim

RUN go install github.com/cosmtrek/air@latest
COPY go.mod go.sum ./
RUN go mod download
COPY ./cmd ./cmd
COPY ./interfaces ./interfaces
COPY ./internal ./internal
COPY ./legacy ./legacy
RUN go build -o exporter ./cmd/exporter.go

FROM golang:1.20.1
WORKDIR /go/src

RUN apt-get update && apt-get install -y python3-pip git

# Use the forked version with support on aggregation pattern
RUN pip3 install 'git+https://github.com/PrasannaSM/mtools.git@develop'
RUN pip3 install python-dateutil
RUN mkdir -p tmp/logs

COPY --from=builder /go/src/exporter ./
CMD ["./exporter"]
