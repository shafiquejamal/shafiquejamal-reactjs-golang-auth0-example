# syntax=docker/dockerfile:1

FROM golang:1.17.1-alpine3.14

WORKDIR /application

COPY server ./server

WORKDIR /application/server

WORKDIR /application
COPY app ./app

RUN apk --no-cache add nodejs yarn --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community && cd app/ && yarn install && yarn build &&\
    cd ../server/ && pwd && ls -al && ls -al cf/ && go get -d && go mod download && \
    CGO_ENABLED=0 go build -ldflags "-w" -a -o react-go . && \
    go get github.com/GeertJohan/go.rice/rice && \
    rice append -i . --exec react-go

EXPOSE 8090    

WORKDIR /application/server
CMD ["./react-go"]