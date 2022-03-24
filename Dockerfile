# Start from golang base image
FROM golang:1.17-alpine as builder
# Set the current working directory inside the container
WORKDIR /app
# Environment
ENV GO_ENV=development \
    APP_ENV=development \
    TZ=Asia/Bangkok
# Install git.
# Git is required for fetching the dependencies.
# RUN apk update && apk add --no-cache git
# Copy go mod and sum files
# COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed
# RUN go mod download
# Copy the source from the current directory to the working Directory inside the container
COPY . .
# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed
RUN go mod download
# Build the Go app
# RUN go build -o storage
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cisrgraderoom_compiler .
### --- End of Builder ---

# Start a new stage from scratch
FROM golang:1.17-alpine
WORKDIR /app
# Install Dependencise
RUN apk add build-base
RUN apk add --no-cache perl-app-cpanminus
RUN apk --update add openjdk8-jre
# Copy the Pre-built binary file from the previous stage. Observe we also copied the .env file
COPY --from=builder /app/cisrgraderoom_compiler .
RUN mkdir -p /app/storage
COPY script /app/script
RUN mkdir -p /app/out
# Command to run the executable
CMD ["./cisrgraderoom_compiler"]


# docker service create --name compiler \
#             --replicas 3 --network cisgraderoom \
#             --env  APP_ENV=production \
#             --env GO_ENV=production \
#             --env  TZ=Asia/Bangkok \
#             --env  MYSQL_ROOT_USERNAME=root \
#             --env  MYSQL_ROOT_PASSWORD=cisgraderoom \
#             --env  MYSQL_CONTAINER=cismariadb \
#             --env MYSQL_DATABASE=cisgraderoom \
#             --env  RABBITMQ_DEFAULT_USER=cisgraderoomcloud \
#             --env  RABBITMQ_DEFAULT_PASS=cisgraderoom \
#             --env  APP_RABBITMQ_VHOST=judge \
#             --env  CIS_RABBITMQ_PROTOCAL=amqp \
#             --env  CIS_RABBITMQ_HOST=cisrabbitmq \
#             --env  CIS_RABBITMQ_PORT=5672 \
#             --env  APP_SERVICE=judge \
#             compiler:latest
