FROM alpine:latest

ARG ADMIN_PORT

WORKDIR /app

COPY ./bin/linux .

EXPOSE $ADMIN_PORT

CMD ["./proxy"]
