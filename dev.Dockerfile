# Start from golang base image
FROM golang:1.17-alpine as builder
WORKDIR /app
RUN apk add build-base
RUN go get github.com/cespare/reflex
COPY . .
ENTRYPOINT ["reflex", "-c", "reflex.conf"]
