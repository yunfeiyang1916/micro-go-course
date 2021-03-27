package endpoint

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"micro-go-course/user-server/service"
)

// 用户终端，负责接收请求，处理请求，并返回结果。可以添加熔断、日志、限流、负载均衡等能力
type UserEndpoints struct {
	// 注册终端
	RegisterEndpoint endpoint.Endpoint
	// 登录终端
	LoginEndpoint endpoint.Endpoint
}

// 登录请求
type LoginReq struct {
	Email    string
	Password string
}

// 登录响应
type LoginResp struct {
	UserInfo *service.UserInfoDTO `json:"user_info"`
}

func MakeLoginEndpoint(userService service.UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*LoginReq)
		userInfo, err := userService.Login(ctx, req.Email, req.Password)
		return &LoginResp{UserInfo: userInfo}, err
	}
}

type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

type RegisterResponse struct {
	UserInfo *service.UserInfoDTO `json:"user_info"`
}

func MakeRegisterEndpoint(userService service.UserService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*RegisterRequest)
		userInfo, err := userService.Register(ctx, &service.RegisterUserVO{
			Username: req.Username,
			Password: req.Password,
			Email:    req.Email,
		})
		return &RegisterResponse{UserInfo: userInfo}, err

	}
}
