# Start from the latest golang base image
FROM golang:latest as builder

ARG GITHUB_USER
ARG GITHUB_TOKEN

# Add Maintainer Info
LABEL maintainer="3almadmoon <info@3almadmoon.com>"

# Set the Current Working Directory inside the container
WORKDIR /app

RUN  git config --global url."https://${GITHUB_USER}:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# Copy go mod and sum files
COPY go.mod go.sum Makefile ./

# Download all dependencies
RUN CGO_ENABLED=0 GOOS=linux make init

COPY . .

RUN CGO_ENABLED=0 GOOS=linux make build-server

######## Start a new stage from scratch #######
FROM alpine:latest

RUN apk update && apk add --no-cache ca-certificates bash git

WORKDIR /app/

COPY config/config.prod.yml config/config.dev.yml config/config.test.yml ./config/

# copy health check binary
COPY bin/grpc_health_probe  .
RUN  chmod +x grpc_health_probe

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/bin/mailboxsvc .
RUN  chmod +x mailboxsvc

# Expose GRPC PORT
EXPOSE 50060

# Command to run the executable
CMD ["./mailboxsvc"]