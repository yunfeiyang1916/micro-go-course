package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/yunfeiyang1916/micro-go-course/grpc-kit/pb"
	"google.golang.org/grpc"

	"golang.org/x/time/rate"

	"github.com/yunfeiyang1916/micro-go-course/grpc-kit/user"
)

func main() {
	flag.Parse()
	var (
		//logger = log.NewLogfmtLogger(os.Stderr)
		//logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		//logger = log.With(logger, "caller", log.DefaultCaller)
		ctx = context.Background()
		// 服务
		svc = user.UserServiceImpl{}
		// 建立endpoint
		endpoint = user.MakeUserEndpoint(svc)
		// 构造限流中间件
		ratebucket = rate.NewLimiter(rate.Every(time.Second*1), 100)
	)
	endpoint = user.NewTokenBucketLimitterWithBuildIn(ratebucket)(endpoint)

	endpts := user.Endpoints{
		UserEndpoint: endpoint,
	}
	// 使用transport构造UserServiceServer
	handler := user.NewUserServer(ctx, endpts)
	// 监听端口，建立gRPC网络服务器，注册RPC服务
	ls, err := net.Listen("tcp", "127.0.0.1:1234")
	if err != nil {
		fmt.Println("Listen error:", err)
		return
	}
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, handler)
	grpcServer.Serve(ls)
}
