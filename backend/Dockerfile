FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o filelocker cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/filelocker .
COPY --from=builder /app/configs ./configs
EXPOSE 9010 9011
CMD ["./filelocker"]