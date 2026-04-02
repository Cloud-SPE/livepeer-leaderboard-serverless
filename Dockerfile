# Use the official Golang image as the base image
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files for dependency management
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy the entire project to the working directory
COPY . .

# Run tests as a non-root user — embedded-postgres (initdb) refuses to run as root.
# The -p 1 flag is required because packages share an embedded Postgres instance.
RUN useradd -m -u 1001 testrunner && chown -R testrunner /app
USER testrunner
RUN go test -p 1 ./...
USER root

# Build the Go app and output it to the 'app' binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o serverless-api main.go

# Ensure entrypoint.sh is executable
RUN chmod +x /app/serverless-api

# Use a minimal base image to run the application (Alpine Linux)
FROM alpine:3.18
ENV PATH=$PATH:/app/
# Set working directory
WORKDIR /app
# Copy the compiled binary from the builder stage
COPY --from=builder /app/serverless-api ./
COPY --from=builder /app/assets/ ./assets/


# Run the binary
CMD ["serverless-api"]