FROM golang:latest
WORKDIR /root/micro-go-course/user-server
COPY / /root/micro-go-course/user-server
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go build -o user-server
EXPOSE 10086
ENTRYPOINT ["./user-server"]