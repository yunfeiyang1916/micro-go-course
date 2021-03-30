package main

import (
	"fmt"
	"github.com/yunfeiyang1916/micro-go-course/go-rpc/service"
	"log"
	"net/rpc"
)

func main() {
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	stringReq := &service.StringReq{
		A: "A",
		B: "B",
	}
	var reply string
	err=client.Call("StringService.Concat",stringReq,&reply)
	fmt.Printf("StringService.Concat : %s concat %s = %s\n",stringReq.A,stringReq.B,reply)
	if err!=nil{
		log.Fatal("Concat error:",err)
	}
	// 异步的调用方式
	call:=client.Go("StringService.Concat",stringReq,&reply,nil)
	_= <-call.Done
	fmt.Printf("StringService.Concat : %s concat %s = %s\n",stringReq.A,stringReq.B,reply)
}
