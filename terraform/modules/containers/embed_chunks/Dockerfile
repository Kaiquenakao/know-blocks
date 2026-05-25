FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY main.go .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o embed-chunks .

# ── Final image ───────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/embed-chunks /embed-chunks

CMD ["/embed-chunks"]
