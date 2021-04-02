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

	"golang.org/x/time/rate"

	"github.com/afex/hystrix-go/hystrix"

	"github.com/yunfeiyang1916/micro-go-course/hystrix/goods/endpoint"
	"github.com/yunfeiyang1916/micro-go-course/hystrix/goods/service"
	"github.com/yunfeiyang1916/micro-go-course/hystrix/goods/transport"
)

func main() {
	servicePort := flag.Int("service.port", 10086, "service port")

	flag.Parse()

	errChan := make(chan error)

	srv := service.NewGoodsServiceImpl()

	// 限流器
	// 第一个参数代表系统每秒钟向令牌桶中放入多少个令牌，也就是限流器平稳状态下每秒可以允许多少请求通过
	// 第二个参数代表令牌桶的上限或者整体大小，也就是限流器允许多大的瞬时请求流量通过
	limiter := rate.NewLimiter(1, 1)

	endpoints := endpoint.GoodsEndpoints{
		GoodsDetailEndpoint: endpoint.MakeGoodsDetailEndpoint(srv, limiter),
	}
	handler := transport.MakeHttpHandler(context.Background(), &endpoints)
	// 修改断路器最低启动阈值为 4 次
	hystrix.ConfigureCommand("Comments", hystrix.CommandConfig{
		RequestVolumeThreshold: 4,
	})

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
