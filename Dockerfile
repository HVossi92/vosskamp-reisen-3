FROM --platform=linux/amd64 golang:1.23 AS builder

RUN apt-get update && apt-get install -y zip

# Set the working directory
WORKDIR /app

# Copy your source code
COPY . .

# Build command
CMD ["make", "build-docker"]
