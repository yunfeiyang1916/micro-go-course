package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/yunfeiyang1916/micro-go-course/hystrix/comment/service"
)

// 评论终端，负责接收请求，处理请求，并返回结果。可以添加熔断、日志、限流、负载均衡等能力
type CommentsEndpoints struct {
	CommentsListEndpoint endpoint.Endpoint
}

// 评论请求结构体
type CommentsListRequest struct {
	Id string
}

// 评论响应结构体
type CommentsListResponse struct {
	Detail service.CommentListVO `json:"detail"`
	Error  string                `json:"error"`
}

// 创建评论的 Endpoint
func MakeCommentsListEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		println("MakeCommentsListEndpoint")
		req := request.(CommentsListRequest)
		detail, err := svc.GetCommentsList(ctx, req.Id)
		var errString = ""
		if err != nil {
			errString = err.Error()
		}
		return &CommentsListResponse{
			Detail: detail,
			Error:  errString,
		}, nil
	}
}
