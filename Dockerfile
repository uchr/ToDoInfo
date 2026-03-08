FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOFLAGS="-trimpath" go build -ldflags="-s -w" -o todoinfo cmd/cli/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/todoinfo /app/todoinfo
ENTRYPOINT ["/app/todoinfo"]
CMD ["bot"]
