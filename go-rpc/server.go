package main

import (
	"github.com/yunfeiyang1916/micro-go-course/go-rpc/service"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func main() {
	stringService := &service.StringService{}
	rpc.Register(stringService)
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(l, nil)
}
