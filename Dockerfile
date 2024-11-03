ARG REGISTRY
FROM ${REGISTRY}/golang:1.20-buster AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v -o main

FROM ${REGISTRY}/debian:buster-slim
WORKDIR /app
RUN update-ca-certificates
COPY --from=builder /app/main ./
CMD ["./main"]