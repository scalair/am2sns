FROM golang:1.13-stretch AS builder
ENV GO111MODULE=on CGO_ENABLED=1
WORKDIR /build
COPY go.mod .
COPY go.sum .
COPY am2sns.go .
RUN go mod download
RUN go build github.com/scalair/am2sns
# https://dev.to/ivan/go-build-a-minimal-docker-image-in-just-three-steps-514i
RUN ldd am2sns | tr -s '[:blank:]' '\n' | grep '^/' | xargs -I % sh -c 'mkdir -p $(dirname ./%); cp % ./%;'
RUN mkdir -p lib64 && cp /lib64/ld-linux-x86-64.so.2 lib64/

FROM scratch
COPY --chown=0:0  --from=builder /build /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/am2sns"]