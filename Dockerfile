# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o main .

# Stage 2: Create a minimal final image
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

# Command to run the executable
CMD ["./main"]
