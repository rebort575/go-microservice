FROM alpine:latest

RUN mkdir /app

COPY AuthApp /app

CMD ["/app/AuthApp"]
