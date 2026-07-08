FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY pkg/ ./pkg/
COPY services/api-gateway/ ./services/api-gateway/

WORKDIR /app/services/api-gateway
RUN go mod download && go mod verify
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /server /server

EXPOSE 8080
CMD ["/server"]
