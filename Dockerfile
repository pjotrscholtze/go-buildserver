# Build
FROM golang:alpine as builder
RUN mkdir /app
WORKDIR /app
ADD go.mod .
RUN go mod download
ADD . /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -o go-buildserver ./cmd/go-buildserver/main.go

RUN find /

# Run
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/go-buildserver .
Expose 8000
CMD ["./go-buildserver"]