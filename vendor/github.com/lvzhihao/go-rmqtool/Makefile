NAME:=gormqtool-test-rabbit
PORT:=6672
MPORT:=16672

stop-docker:
	@if test "$(shell sudo docker ps -a | grep $(NAME) | wc -l)" = "1"; then \
		echo "docker stop" && \
		sudo docker stop $(shell sudo docker ps -a | grep $(NAME) | awk '{print $$1}'); \
	fi

force-stop-docker: 
	sudo docker rm $(NAME)

start-docker: stop-docker
	echo "docker start" && sudo docker run -d --rm --hostname test-rabbit --name $(NAME) -p $(MPORT):15672 -p $(PORT):5672 rabbitmq:3-management

start-test-rmq: start-docker

stop-test-rmq: stop-docker

test:
	go test -v
