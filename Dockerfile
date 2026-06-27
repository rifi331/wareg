# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/wareg .

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /app

# Copy the compiled binary plus the runtime assets the server serves.
# The original image only copied frontend/, so /static/* 404'd in the container.
COPY --from=builder /out/wareg /app/wareg
COPY --from=builder /app/frontend /app/frontend
COPY --from=builder /app/static /app/static

ENV PORT=7001
EXPOSE 7001

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:${PORT:-7001}/healthz || exit 1

CMD ["/app/wareg"]
