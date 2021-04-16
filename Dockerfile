FROM golang:alpine as builder

WORKDIR /go/src/github.com/proffust/huawei-perf
COPY . .
RUN apk --no-cache add git
RUN go build

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /

COPY --from=builder /go/src/github.com/proffust/huawei-perf/huawei-perf ./usr/bin/huawei-perf

CMD ["huawei-perf", "-config", "/etc/config.yml"]
