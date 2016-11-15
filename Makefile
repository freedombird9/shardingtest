BUILD_NUMBER?=latest

install:
	go install redisshardingtest/...

deps:
	-cd $(GOPATH)/src; \
	if [ ! -d "go-jasperlib" ]; then git clone http://qa1-sjc002-030.i.jasperwireless.com/cc/go-jasperlib.git; fi
	-go get -t -u -f -insecure redisshardingtest/...

docker:
	docker build  -t redisshardingtest:$(BUILD_NUMBER) .

FORCE:
.PHONY: deps docker
