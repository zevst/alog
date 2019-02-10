FROM golang:1.11.5-alpine3.9

WORKDIR /go/src/alog
COPY . .

RUN go install -v
RUN go build .

CMD ["./alog"]