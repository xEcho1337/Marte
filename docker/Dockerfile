FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY marte /usr/bin/marte
COPY frontend/ /frontend/
WORKDIR /app
VOLUME /app
EXPOSE 14100 14101
ENTRYPOINT ["marte"]
