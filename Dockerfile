FROM golang
MAINTAINER Yongfeng Zhang <yongfezh@cisco.com> 
ADD . /go/src/redisshardingtest
WORKDIR /go/src/redisshardingtest
RUN make deps
RUN make install
WORKDIR /go/bin
EXPOSE 8080
CMD ["bash"]
