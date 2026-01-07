# Use the official Go image for development
FROM golang:1.23

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first and download modules
COPY go.mod go.sum ./
RUN go mod download

# Expose the app port (matches APP_PORT in .env)
EXPOSE 8080

# Default command (overridden by docker-compose)
CMD ["go", "run", "cmd/main.go"]
