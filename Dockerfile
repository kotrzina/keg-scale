FROM golang:1.22 AS backend
ENV CGO_ENABLED 0
ADD backend /app
WORKDIR /app
RUN go build -ldflags "-s -w" -v -o keg-scale .


FROM node:18 AS  frontend
ADD frontend /app
WORKDIR /app
RUN yarn
RUN yarn build


FROM alpine:3
RUN apk update && \
    apk add openssl tzdata && \
    rm -rf /var/cache/apk/* \
    && mkdir /app

WORKDIR /app

ADD Dockerfile /Dockerfile
ENV FRONTEND_PATH /app/frontend/

COPY --from=backend /app/keg-scale /app/keg-scale
COPY --from=frontend /app/build/ /app/frontend/

RUN chown nobody /app/keg-scale \
    && chmod 500 /app/keg-scale

USER nobody

ENTRYPOINT ["/app/keg-scale"]