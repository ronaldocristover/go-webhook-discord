# Start with the official Golang image
FROM golang:1.22-alpine

# Set environment variable for Go
ENV GO111MODULE=on

# Create and set the working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Install dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o webhook-service .

# Expose the port on which the service will run
EXPOSE 8080

# Start the application
CMD ["./webhook-service"]