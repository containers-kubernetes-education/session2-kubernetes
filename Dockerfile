############################
# STEP 1 build executable binary
############################
FROM golang AS builder
WORKDIR $GOPATH/src/github.com/containers-kubernetes-education/session2-kubernetes
COPY . /go/src/github.com/containers-kubernetes-education/session2-kubernetes

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=auto
ENV GOPATH=/go

WORKDIR $GOPATH/src/github.com/containers-kubernetes-education/session2-kubernetes
COPY . /go/src/github.com/containers-kubernetes-education/session2-kubernetes

RUN go build -o /go/bin/run /go/src/github.com/containers-kubernetes-education/session2-kubernetes/cmd/main.go

############################
# STEP 2 build a small image
############################
FROM scratch
COPY --from=builder /go/bin/run /go/bin/run
COPY assets /assets
COPY data/names.json /data/names.json
COPY config/defaults.json /config
ENTRYPOINT ["/go/bin/run"]