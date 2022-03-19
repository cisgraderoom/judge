# Start from golang base image
FROM golang:1.17-alpine as builder
WORKDIR /app
RUN apk add build-base
RUN apk add --no-cache perl-app-cpanminus
RUN apk --update add openjdk8-jre
RUN go get github.com/cespare/reflex
COPY . .
ENTRYPOINT ["reflex", "-c", "reflex.conf"]
