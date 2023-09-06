FROM golang:1.18-buster AS build

WORKDIR /app

COPY ./ ./

WORKDIR /app/spec
RUN go mod download

WORKDIR /app/container5
RUN go mod download

WORKDIR /app/container5/app
RUN go mod tidy
RUN go build -o /container5
FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=build container5 container5
ENTRYPOINT ["/container5"]

