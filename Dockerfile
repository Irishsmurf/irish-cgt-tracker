# Change from 1.23-alpine to 1.24-alpine
FROM golang:1.24-alpine AS builder

# ... rest of the file remains exactly the same ...
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cgt-tracker main.go

FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=builder /app/cgt-tracker .
COPY --from=builder /app/web ./web
EXPOSE 8080
VOLUME ["/app/data"]
CMD ["./cgt-tracker"]
