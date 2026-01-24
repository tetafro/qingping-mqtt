FROM golang:1.24-alpine3.21 AS build

WORKDIR /build

RUN apk add --no-cache git gcc musl-dev

COPY . .

RUN go build -o ./bin/qingping-mqtt .

FROM alpine:3.21

WORKDIR /app

COPY --from=build /build/bin/qingping-mqtt /app/

RUN apk add --no-cache ca-certificates && \
    addgroup -S -g 5000 app && \
    adduser -S -u 5000 -G app app && \
    chown -R app:app .

USER app
EXPOSE 8080 1883

CMD ["/app/qingping-mqtt"]
