# Deferred for the current phase.
# Containerization can be re-enabled when we add a deployment pipeline.
#
# FROM golang:1.25-alpine AS build
# WORKDIR /app
# COPY go.mod ./
# COPY src ./src
# RUN go build -o /bin/image-compressar ./src/cmd/api
#
# FROM alpine:3.22
# WORKDIR /app
# COPY --from=build /bin/image-compressar /usr/local/bin/image-compressar
# EXPOSE 8080
# CMD ["image-compressar"]
