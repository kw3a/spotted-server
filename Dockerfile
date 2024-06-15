FROM --platform=linux/amd64 debian:stable-slim

RUN apt-get update && apt-get install -y ca-certificates

ADD cmd/api/out /usr/bin/out

CMD ["out"]