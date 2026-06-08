FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ecowitt-prom .

FROM alpine:3.21
RUN apk --no-cache add ca-certificates && adduser -D appuser
COPY --from=builder /app/ecowitt-prom /usr/local/bin/ecowitt-prom
USER appuser
EXPOSE 8080
ENTRYPOINT ["ecowitt-prom"]