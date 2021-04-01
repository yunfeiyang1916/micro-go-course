package main

import (
	"context"
	"fmt"

	"github.com/yunfeiyang1916/micro-go-course/grpc/pb"
	"google.golang.org/grpc"
)

func main() {
	serviceAddress := "127.0.0.1:1234"
	conn, err := grpc.Dial(serviceAddress, grpc.WithInsecure())
	if err != nil {
		panic("connect error")
	}
	defer conn.Close()

	userClient := pb.NewUserServiceClient(conn)
	userReq := &pb.LoginReq{Username: "admin", Password: "admin"}
	reply, _ := userClient.CheckPassword(context.Background(), userReq)
	fmt.Printf("UserService CheckPassword : %s", reply.Ret)
}
