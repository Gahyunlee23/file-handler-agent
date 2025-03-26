# Build stage
FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /file-handler-agent ./cmd/server

# Final stage
FROM alpine:3.16

# install Ghostscript
RUN apk add --no-cache ghostscript

# work dir
WORKDIR /app
RUN mkdir -p /app/temp/input /app/temp/output

# copy binery
COPY --from=builder /file-handler-agent /app/

# expose port
EXPOSE 8080

# execute
CMD ["/app/file-handler-agent"]