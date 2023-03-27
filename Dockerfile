# Build
FROM golang:1.20-alpine as builder
RUN mkdir /app
WORKDIR /app
ADD go.mod .
RUN go mod download
ADD . /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -o go-buildserver ./cmd/go-buildserver/main.go

# Run
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/go-buildserver .
COPY ./cmd/go-buildserver/templates ./templates
COPY ./example ./example
COPY ./entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

Expose 3000
RUN apk add --update --no-cache openssh git
CMD ["./entrypoint.sh"]