FROM golang:1.9

WORKDIR /go/src/github.com/lvzhihao/zhiya

COPY . .  

RUN go-wrapper install \
    && rm -rf *
