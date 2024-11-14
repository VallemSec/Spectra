# Start from the latest golang base image for building the application
FROM golang:1.23-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Install git and other dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main .

# Start a new stage from a smaller base image
FROM alpine:latest

# Install necessary dependencies for the Go binary
RUN apk add --no-cache ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy all the files from the previous stage
COPY --from=builder /app/main .

# Copy the config file
COPY config.yaml .

# Ensure the binary has executable permissions
RUN chmod +x ./main

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
