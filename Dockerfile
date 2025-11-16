FROM golang:1.24-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org
COPY . .
RUN go build -o /fleet cmd/backend/main.go
RUN go build -o /worker worker/worker.go

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /fleet /fleet
EXPOSE 8080
ENTRYPOINT ["/fleet"]
