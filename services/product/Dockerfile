FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY pkg/ ./pkg/
COPY services/product/ ./services/product/

WORKDIR /app/services/product
RUN go mod download && go mod verify
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /server /server

EXPOSE 8082
CMD ["/server"]
