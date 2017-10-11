OS := $(shell uname)

build: */*.go
	go build 

api: build
	./zhiya api

docker-build:
	sudo docker build -t edwinlll/zhiya:latest .

docker-push:
	sudo docker push edwinlll/zhiya:latest
