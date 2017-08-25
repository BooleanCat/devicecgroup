FROM ubuntu:zesty

RUN apt-get update && apt-get install -y git

ADD https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz .
RUN tar -C /usr/local -xzf go1.8.3.linux-amd64.tar.gz
RUN rm go1.8.3.linux-amd64.tar.gz
ENV PATH "${PATH}:/usr/local/go/bin"
ENV GOPATH "/root/go"
ENV PATH "${PATH}:${GOPATH}/bin"

RUN go get github.com/onsi/ginkgo/ginkgo
RUN go get github.com/onsi/gomega
