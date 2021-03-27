package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/yunfeiyang1916/micro-go-course/user-server/dao"
	"github.com/yunfeiyang1916/micro-go-course/user-server/endpoint"
	"github.com/yunfeiyang1916/micro-go-course/user-server/redis"
	"github.com/yunfeiyang1916/micro-go-course/user-server/service"
	"github.com/yunfeiyang1916/micro-go-course/user-server/transport"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// 依次组建 service、endpoint 和 transport，并启动 Web 服务器
func main() {
	var (
		// 服务监听端口
		servicePort = flag.Int("service.port", 10086, "service port")
	)
	flag.Parse()

	// 使用k8s部署时，延时启动，等待 MySQL 和 Redis 准备好
	time.Sleep(10 * time.Second)

	ctx := context.Background()
	errChan := make(chan error)
	err := dao.InitMysql("localhost", "3306", "root", "root123456", "user")
	if err != nil {
		log.Fatal(err)
	}
	err = redis.InitRedis("localhost", "6379", "")
	if err != nil {
		log.Fatal(err)
	}

	userService := service.MakeUserServiceImpl(&dao.UserDAOImpl{})

	userEndpoints := &endpoint.UserEndpoints{
		RegisterEndpoint: endpoint.MakeRegisterEndpoint(userService),
		LoginEndpoint:    endpoint.MakeLoginEndpoint(userService),
	}
	r := transport.MakeHttpHandler(ctx, userEndpoints)

	go func() {
		errChan <- http.ListenAndServe(":"+strconv.Itoa(*servicePort), r)
	}()

	go func() {
		// 监控系统信号，等待ctrl+c系统信号通知服务关闭
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
	error := <-errChan
	log.Println(error)
}
