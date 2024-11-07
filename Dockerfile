# First stage: Build the Go binary
FROM --platform=linux/amd64 golang:1.21 as builder

# Set the working directory in the Go build container
WORKDIR /app

# Copy all Go files into the container
COPY . .

# Ensure the Go files are in the correct directory and build the Go binary
RUN go build -o /app/out ./cmd/api

CMD ["./out"]
