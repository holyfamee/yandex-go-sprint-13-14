FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o scheduler .

FROM alpine:latest

RUN apk --no-cache add ca-certificates libc6-compat

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /app/scheduler .
COPY --from=builder /app/web ./web

RUN mkdir -p /app/data && chown -R appuser:appuser /app

USER appuser

ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/data/scheduler.db

EXPOSE 7540

CMD ["./scheduler"]
