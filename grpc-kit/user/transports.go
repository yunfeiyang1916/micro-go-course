package user

import (
	"context"
	"errors"

	"github.com/yunfeiyang1916/micro-go-course/grpc-kit/pb"

	"github.com/go-kit/kit/transport/grpc"
)

var (
	ErrorBadRequest = errors.New("invalid request parameter")
)

type grpcServer struct {
	checkPassword grpc.Handler
}

func NewUserServer(ctx context.Context, endpoints Endpoints) pb.UserServiceServer {
	return &grpcServer{
		checkPassword: grpc.NewServer(endpoints.UserEndpoint, DecodeLoginRequest, EncodeLoginResponse),
	}
}

// 实现grpc接口
func (g *grpcServer) CheckPassword(ctx context.Context, r *pb.LoginReq) (*pb.LoginResp, error) {
	// 借用grpc
	_, resp, err := g.checkPassword.ServeGRPC(ctx, r)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.LoginResp), nil
}

// 解码请求体
func DecodeLoginRequest(ctx context.Context, r interface{}) (interface{}, error) {
	req := r.(*pb.LoginReq)
	return LoginForm{
		Username: req.Username,
		Password: req.Password,
	}, nil
}

// 编码响应结果
func EncodeLoginResponse(_ context.Context, r interface{}) (interface{}, error) {
	resp := r.(LoginResult)
	retStr := "fail"
	if resp.Ret {
		retStr = "success"
	}
	errStr := ""
	if resp.Err != nil {
		errStr = resp.Err.Error()
	}
	return &pb.LoginResp{
		Ret: retStr,
		Err: errStr,
	}, nil
}
