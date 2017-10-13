OS := $(shell uname)

build: */*.go
	go build 

api: build
	./zhiya api

docker-build:
	sudo docker build -t edwinlll/zhiya:latest .

docker-push:
	sudo docker push edwinlll/zhiya:latest

docker-ccr:
	sudo docker tag edwinlll/zhiya:latest ccr.ccs.tencentyun.com/wdwd/zhiya
	sudo docker push ccr.ccs.tencentyun.com/wdwd/zhiya
