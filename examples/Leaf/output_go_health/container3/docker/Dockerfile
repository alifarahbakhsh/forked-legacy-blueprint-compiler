FROM golang:1.18-buster AS build

WORKDIR /app

COPY ./ ./

WORKDIR /app/spec
RUN go mod download

WORKDIR /app/container3
RUN go mod download

WORKDIR /app/container3/app
RUN go mod tidy
RUN go build -o /container3
FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=build container3 container3
ENTRYPOINT ["/container3"]

