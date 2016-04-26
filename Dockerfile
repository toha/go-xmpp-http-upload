FROM golang

ADD . /go/src/git.tha.io/toha/go-xmpp-upload

RUN go get github.com/lib/pq
RUN go install git.tha.io/toha/go-xmpp-upload

ENTRYPOINT /go/bin/go-xmpp-upload

EXPOSE 8080
