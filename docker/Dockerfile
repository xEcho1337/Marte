FROM alpine:3.21
COPY ca-certificates.crt /etc/ssl/certs/
COPY marte /usr/bin/marte
COPY frontend/ /frontend/
WORKDIR /app
VOLUME /app
EXPOSE 14100 14101
ENTRYPOINT ["marte"]
