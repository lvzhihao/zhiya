FROM golang:1.9.1

COPY . /go/src/github.com/lvzhihao/zhiya 

WORKDIR /go/src/github.com/lvzhihao/zhiya

RUN rm -f /go/src/github.com/lvzhihao/zhiya/.zhiya.yaml
RUN go-wrapper install
