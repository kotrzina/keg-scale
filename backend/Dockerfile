FROM golang:1.22 AS builder
ENV CGO_ENABLED 0
ADD . /app
WORKDIR /app
RUN go build -ldflags "-s -w" -v -o keg-scale .

FROM alpine:3
RUN apk update && \
    apk add openssl tzdata && \
    rm -rf /var/cache/apk/* \
    && mkdir /app

WORKDIR /app

ADD Dockerfile /Dockerfile

COPY --from=builder /app/keg-scale /app/keg-scale

RUN chown nobody /app/keg-scale \
    && chmod 500 /app/keg-scale

USER nobody

ENTRYPOINT ["/app/keg-scale"]