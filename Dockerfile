FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /eform-api ./cmd/api

FROM alpine:3.22

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /eform-api /app/eform-api
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/seeds /app/seeds

RUN mkdir -p /app/uploads && chown -R app:app /app

USER app

EXPOSE 8080

CMD ["/app/eform-api"]
