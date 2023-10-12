FROM golang:1.21

WORKDIR /usr/src/app

COPY go.* ./

RUN go mod download

COPY ./ ./
