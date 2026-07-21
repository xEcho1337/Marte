FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY backend/ ./backend/
COPY shared/ ./shared/
COPY client/ ./client/
WORKDIR /app/backend
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /marte .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /marte /usr/bin/marte
COPY frontend/ /frontend/
WORKDIR /app
VOLUME /app
EXPOSE 14100 14101
ENTRYPOINT ["marte"]
