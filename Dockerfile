FROM golang:1.23

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

# Set up environment variables for cross-compilation
ENV GOOS=linux
ENV GOARCH=amd64

EXPOSE 8080

RUN go build -o cmd/api/main ./cmd/api/main.go

CMD ["./cmd/api/main"]