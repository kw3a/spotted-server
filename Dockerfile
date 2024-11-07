FROM --platform=linux/amd64 golang:1.22.4 as builder

RUN apt-get update && apt-get install -y ca-certificates

WORKDIR /app

COPY ./views /app/views
ADD cmd/api/out /app/out

CMD ["./out"]
