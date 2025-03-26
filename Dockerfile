FROM golang:1.23-alpine

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git build-base ghostscript

# Install air
RUN go install github.com/air-verse/air@latest

# Prepare directories
RUN mkdir -p /app/temp/input /app/temp/output

# Copy Go modules first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# expose port
EXPOSE 7701

# execute with air
CMD ["air", "-c", ".air.toml"]