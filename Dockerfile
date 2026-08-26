# Build stage
FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Final stage
FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/server .

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/app/server"]