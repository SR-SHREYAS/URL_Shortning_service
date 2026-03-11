FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/main .

FROM alpine:3.20

RUN adduser -S -D -H -h /app appuser

WORKDIR /app

COPY --from=builder /out/main ./main
COPY public ./public

USER appuser

EXPOSE 4000

CMD ["./main"]

