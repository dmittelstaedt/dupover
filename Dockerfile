FROM golang:1.11
ARG http_proxy
ARG https_proxy
ARG os=linux
ARG arch=amd64

ENV http_proxy=${http_proxy}
ENV https_proxy=${https_proxy}

RUN git clone https://github.com/dmittelstaedt/dupover.git /go/src/app

WORKDIR /go/src/app

RUN go get ./...
RUN VERSION=$(git tag --list | tail -1 | cut -c 2-) && \
GIT_COMMIT=$(git rev-parse --short HEAD) && \
BUILD_DATE=$(date +"%Y-%m-%d %T") && \
GOOS=${os} GOARCH=${arch} go build -ldflags "-X main.versionNumber=$VERSION -X main.gitCommit=$GIT_COMMIT -X 'main.buildDate=$BUILD_DATE'" -o dupover main.go

ENTRYPOINT [ "/bin/bash" ]
