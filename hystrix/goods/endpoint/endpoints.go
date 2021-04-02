package endpoint

import (
	"context"
	"errors"

	"golang.org/x/time/rate"

	"github.com/go-kit/kit/endpoint"
	"github.com/yunfeiyang1916/micro-go-course/hystrix/goods/service"
)

type GoodsEndpoints struct {
	GoodsDetailEndpoint endpoint.Endpoint
}

// 商品详情请求结构体
type GoodsDetailRequest struct {
	Id string
}

// 商品详情响应结构体
type GoodsDetailResponse struct {
	Detail service.GoodsDetailVO `json:"detail"`
	Error  string                `json:"error"`
}

// 创建商品详情的 Endpoint
func MakeGoodsDetailEndpoint(svc service.Service, limiter *rate.Limiter) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		// 判断是否被限流
		if !limiter.Allow() {
			// Allow返回false，表示桶内不足一个令牌，应该被限流，默认返回 ErrLimiExceed 异常
			return nil, errors.New("ErrLimitExceed")
		}
		req := request.(GoodsDetailRequest)
		detail, err := svc.GetGoodsDetail(ctx, req.Id)
		var errString = ""

		if err != nil {
			errString = err.Error()
			return &GoodsDetailResponse{
				Detail: detail,
				Error:  errString,
			}, nil
		}
		return &GoodsDetailResponse{
			Detail: detail,
			Error:  errString,
		}, nil
	}
}
