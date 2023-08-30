FROM golang:1.18-buster AS build

WORKDIR /app

COPY ./ ./

WORKDIR /app/spec
RUN go mod download

WORKDIR /app/container4
RUN go mod download

WORKDIR /app/container4/app
RUN go mod tidy
RUN go build -o /container4
FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=build container4 container4
ENTRYPOINT ["/container4"]

