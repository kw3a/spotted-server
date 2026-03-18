FROM alpine:latest

RUN apk add --no-cache curl

WORKDIR /app

COPY ./views /app/views
COPY ./public /app/public
ADD cmd/api/bin /app/bin

CMD ["./bin"]
