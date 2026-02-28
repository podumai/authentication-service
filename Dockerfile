FROM golang:1.26.0-alpine3.23 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main ./cmd/main.go

FROM alpine:3.23 AS production

WORKDIR /app

COPY --from=build /app/main .

EXPOSE 8080
CMD ["./main"]

FROM golang:1.26.0-alpine3.23 AS development

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/air-verse/air@v1.52.3

COPY . .

EXPOSE 8080
EXPOSE 8081
CMD ["air", "-c", ".air.toml"]