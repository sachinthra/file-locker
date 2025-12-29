FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o filelocker cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /app/filelocker .
ARG SERVER_PORT=9010
ARG GRPC_PORT=9011
EXPOSE ${SERVER_PORT} ${GRPC_PORT}
CMD ["./filelocker"]