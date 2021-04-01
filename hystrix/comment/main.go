package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/yunfeiyang1916/micro-go-course/hystrix/comment/endpoint"
	"github.com/yunfeiyang1916/micro-go-course/hystrix/comment/service"
	"github.com/yunfeiyang1916/micro-go-course/hystrix/comment/transport"
)

func main() {
	servicePort := flag.Int("service.port", 10086, "service port")

	flag.Parse()

	errChan := make(chan error)

	srv := service.NewGoodsServiceImpl()

	endpoints := endpoint.CommentsEndpoints{
		CommentsListEndpoint: endpoint.MakeCommentsListEndpoint(srv),
	}

	handler := transport.MakeHttpHandler(context.Background(), &endpoints)

	go func() {
		errChan <- http.ListenAndServe(":"+strconv.Itoa(*servicePort), handler)
	}()

	go func() {
		// 监控系统信号，等待 ctrl + c 系统信号通知服务关闭
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	err := <-errChan
	log.Printf("listen err : %s", err)
}
