FROM golang:1.19.4-alpine
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./models ./models
COPY ./restapi ./restapi
COPY ./cmd ./cmd
RUN go build -o /go-buildserver ./cmd/go-buildserver
