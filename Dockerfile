FROM golang:1.22-alpine AS builder

RUN apk add --no-cache build-base libwebp-dev

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies first
# This leverages Docker's layer caching to avoid re-downloading dependencies on every build
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application, creating a statically linked binary
# This allows the binary to run in a minimal container without external dependencies
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /e-oasis ./cmd/main.go
# https://github.com/chai2010/webp need open cgo
RUN CGO_ENABLED=1 GOOS=linux go build -o /e-oasis ./cmd/main.go

FROM alpine:3.22

RUN apk add --no-cache libwebp

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /e-oasis /app/e-oasis

# Copy non-code assets required at runtime
COPY ./internal/templates/ ./templates/
COPY ./internal/static/ ./static/

# The application uses this directory to store data (databases, books, covers, etc.)
# We declare it as a volume so data can be persisted outside the container.
VOLUME ["/var/opt/e-oasis"]

# Expose the port the application listens on
EXPOSE 8080

# The command to run when the container starts
ENTRYPOINT ["/app/e-oasis"]
