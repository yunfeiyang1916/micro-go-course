package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yunfeiyang1916/micro-go-course/grpc-kit/pb"
	"google.golang.org/grpc"
)

func main() {
	serviceAddr := "127.0.0.1:1234"
	conn, err := grpc.Dial(serviceAddr, grpc.WithInsecure())
	if err != nil {
		panic("connect error")
	}
	defer conn.Close()
	userClient := pb.NewUserServiceClient(conn)
	stringReq := &pb.LoginReq{Username: "admin", Password: "admin"}
	reply, err := userClient.CheckPassword(context.Background(), stringReq)
	if err != nil {
		log.Fatal("userClient.CheckPassword error:", err)
	}
	fmt.Printf("CheckPassword ret is %s\n", reply.Ret)
}
