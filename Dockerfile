FROM golang:1.12.5-alpine as builder

WORKDIR /go/src/app
COPY Gopkg.lock Gopkg.toml monohook.go /go/src/app/

RUN apk add --no-cache curl git && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
    dep ensure && \
    go build

FROM alpine:3.9

LABEL maintainer="npm@alphahydrae.com"

COPY --from=builder /go/src/app/app /usr/local/bin/monohook

ENTRYPOINT [ "/usr/local/bin/monohook" ]
