FROM golang:alpine AS builder

RUN mkdir /build

ADD . /build/

WORKDIR /build

RUN go build -o main .

#stage 2

FROM alpine

RUN adduser -S -D -H -h /app appuser

USER appuser

COPY --from=builder /build/main /app/main

WORKDIR /app

EXPOSE 3000

CMD ["./main"]