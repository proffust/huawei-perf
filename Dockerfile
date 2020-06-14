FROM golang:alpine as builder

RUN apk --no-cache add git
RUN go get -u github.com/proffust/huawei-perf

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /

COPY --from=builder /go/bin/huawei-perf ./usr/bin/huawei-perf

CMD ["huawei-perf", "-config", "/etc/config.yml"]
