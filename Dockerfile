FROM golang:1.12

ENV GO111MODULE=on

RUN mkdir -p /go/src/github.com/tukejonny/raftlock
WORKDIR /go/src/github.com/tukejonny/raftlock

COPY . .

RUN GOOS=linux GOARCH=amd64 go install .
RUN ls -l /go/bin/raftlock

RUN mkdir -p /usr/src/app
RUN mkdir -p /usr/src/app/raftdir

ENTRYPOINT ["/go/bin/raftlock"]