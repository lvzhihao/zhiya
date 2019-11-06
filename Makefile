OS := $(shell uname)

build: */*.go
	go build 

api: build
	./zhiya api

docker-build:
	docker build -t edwinlll/zhiya:latest .

docker-push:
	docker push edwinlll/zhiya:latest

docker-ccr:
	docker tag edwinlll/zhiya:latest ccr.ccs.tencentyun.com/wdwd/zhiya:latest
	docker push ccr.ccs.tencentyun.com/wdwd/zhiya:latest
	docker rmi ccr.ccs.tencentyun.com/wdwd/zhiya:latest

docker-uhub:
	docker tag edwinlll/zhiya:latest uhub.service.ucloud.cn/mmzs/zhiya:latest
	docker push uhub.service.ucloud.cn/mmzs/zhiya:latest
	docker rmi uhub.service.ucloud.cn/mmzs/zhiya:latest

docker-ali:
	docker tag edwinlll/zhiya:latest registry.cn-hangzhou.aliyuncs.com/weishangye/zhiya:latest
	docker push registry.cn-hangzhou.aliyuncs.com/weishangye/zhiya:latest
	docker rmi registry.cn-hangzhou.aliyuncs.com/weishangye/zhiya:latest

docker-wdwd:
	docker tag edwinlll/zhiya:latest docker.wdwd.com/wxsq/zhiya:latest
	docker push docker.wdwd.com/wxsq/zhiya:latest
	docker rmi docker.wdwd.com/wxsq/zhiya:latest
