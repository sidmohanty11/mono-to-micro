FROM alpine:latest

RUN mkdir /app

COPY mailerApp /app
COPY templates /app/templates

WORKDIR /app

CMD ["./mailerApp"]