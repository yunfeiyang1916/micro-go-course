package main

import (
	"flag"
	"log"
	"net"

	"github.com/yunfeiyang1916/micro-go-course/grpc/pb"
	"github.com/yunfeiyang1916/micro-go-course/grpc/user"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()
	l, err := net.Listen("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatalln("Listen error:", err)
	}
	grpcServer := grpc.NewServer()
	useService := &user.UserService{}
	pb.RegisterUserServiceServer(grpcServer, useService)
	grpcServer.Serve(l)
}
