FROM golang:1.10-alpine

WORKDIR /go/src/github.com/UCCNetworkingSociety/Netsoc-Discord-Bot
COPY . .

RUN apk update && \
    apk add --no-cache git && \
    git remote set-url origin https://github.com/UCCNetworkingSociety/Netsoc-Discord-Bot

RUN mkdir /logs
RUN go get -d -v ./...
RUN go install -v ./...

CMD ["Netsoc-Discord-Bot", "-log_dir", "/logs", "-alsologtostderr"]
