# Stage 1: Build the Go application
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./

# Download dependencies. This is a separate step to leverage Docker's layer caching.
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application. CGO_ENABLED=0 is important for creating a static binary.
# The output binary will be named 'server'.
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./main.go

# Stage 2: Create the final, lightweight image
FROM alpine:latest

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server .

# Expose the port the app runs on (Render will override this with the PORT env var)
EXPOSE 4000

# Command to run the executable
CMD ["./server"]