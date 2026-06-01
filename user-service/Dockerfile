FROM golang:1.23-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o user-service ./cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=0 /app/user-service .

EXPOSE 8082
CMD ["./user-service"]
