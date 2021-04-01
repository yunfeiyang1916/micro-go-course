package user

import (
	"context"

	"github.com/yunfeiyang1916/micro-go-course/grpc/pb"
)

// 用户服务
type UserService struct {
}

// 校验密码
func (s *UserService) CheckPassword(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	if req.Username == "admin" && req.Password == "admin" {
		response := pb.LoginResp{Ret: "success"}
		return &response, nil
	}

	response := pb.LoginResp{Ret: "fail"}
	return &response, nil
}
