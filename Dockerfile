FROM golang:1.9 as builder
WORKDIR /go/src/github.com/lvzhihao/zhiya
COPY . . 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /usr/local/zhiya
COPY --from=builder /go/src/github.com/lvzhihao/zhiya/zhiya .
COPY ./docker-entrypoint.sh  .
ENV PATH /usr/local/zhiya:$PATH
RUN chmod +x /usr/local/zhiya/docker-entrypoint.sh
ENTRYPOINT ["/usr/local/zhiya/docker-entrypoint.sh"]
