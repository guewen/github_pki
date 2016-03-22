FROM golang:alpine
RUN apk update && \
    apk add git && \
    go get github.com/camptocamp/github_pki && \
    apk del git
WORKDIR /go/bin
ENTRYPOINT ["github_pki"]
