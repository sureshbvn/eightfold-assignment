# Docker file for logporcessor service
# Author: Suresh Bysani

# Use the official Golang image as the base image
FROM golang:1.17

# Set the working directory inside the container
WORKDIR logprocessor

# Enable Go modules outside of $GOPATH
ENV GO111MODULE=on

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build  -race -o app -v ./cmd/main.go

# Run the Go application when the container starts
CMD ["./app"]