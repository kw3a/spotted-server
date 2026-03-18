FROM alpine:latest

RUN apk add --no-cache curl

WORKDIR /app

COPY ./views /app/views
ADD cmd/api/bin /app/bin

CMD ["./bin"]
