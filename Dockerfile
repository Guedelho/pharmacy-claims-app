FROM golang:1.24-alpine AS base
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
COPY .air.toml ./
RUN go install github.com/air-verse/air@latest
RUN go mod tidy
RUN go mod download
COPY . .
RUN mkdir -p logs
EXPOSE 8080
CMD ["air"]
