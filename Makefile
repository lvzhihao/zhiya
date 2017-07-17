OS := $(shell uname)

build: */*.go
	go build 

api: build
	./zhiya api
