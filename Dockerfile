# base go image
FROM golang:1.23.3-alpine3.20 AS builder


RUN mkdir /app
ADD . /app
WORKDIR /app


# Build the Go application
RUN CGO_ENABLED=0 go build -o trustgo ./cmd
RUN chmod +x /app/trustgo

# Final lightweight image
FROM alpine:latest

# Create a directory for the application
RUN mkdir /app

# Copy the executable from the builder stage to the final image
COPY --from=builder /app/trustgo /app


# Set the working directory
WORKDIR /app

# Command to run the executable
CMD ["./trustgo"]
