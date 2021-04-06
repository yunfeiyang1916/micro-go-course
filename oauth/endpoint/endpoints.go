package endpoint

import (
	"context"
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/yunfeiyang1916/micro-go-course/oauth/model"
	"github.com/yunfeiyang1916/micro-go-course/oauth/service"
	"net/http"

	"github.com/go-kit/kit/endpoint"
)

// 终端层，负责接收请求，处理请求，并返回结果。可以添加熔断、日志、限流、负载均衡等能力
type OAuth2Endpoints struct {
	// 令牌终端
	TokenEndpoint endpoint.Endpoint
	// 校验令牌终端
	CheckTokenEndpoint endpoint.Endpoint
	// 健康检测终端
	HealthCheckEndpoint endpoint.Endpoint
}

// 请求上下文使用的key
const (
	// 认证详情key
	OAuth2DetailsKey = "OAuth2Details"
	// 客户端详情key
	OAuth2ClientDetailsKey = "OAuth2ClientDetails"
	// 认证错误key
	OAuth2ErrorKey = "OAuth2Error"
)

var (
	ErrInvalidClientRequest = errors.New("invalid client message")
	ErrInvalidUserRequest   = errors.New("invalid user message")
	ErrNotPermit            = errors.New("not permit")
)

// 令牌请求
type TokenRequest struct {
	GrantType string
	Reader    *http.Request
}

// 令牌响应
type TokenResponse struct {
	AccessToken *model.OAuth2Token `json:"access_token"`
	Error       string             `json:"error"`
}

type CheckTokenRequest struct {
	Token         string
	ClientDetails model.ClientDetails
}

type CheckTokenResponse struct {
	OAuthDetails *model.OAuth2Details `json:"o_auth_details"`
	Error        string               `json:"error"`
}

// HealthRequest 健康检查请求结构
type HealthRequest struct{}

// HealthResponse 健康检查响应结构
type HealthResponse struct {
	Status bool `json:"status"`
}

// MakeHealthCheckEndpoint 创建健康检查Endpoint
func MakeHealthCheckEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		status := svc.HealthCheck()
		return HealthResponse{
			Status: status,
		}, nil
	}
}

// 创建认证中间件
func MakeClientAuthorizationMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if err, ok := ctx.Value(OAuth2ErrorKey).(error); ok {
				return nil, err
			}
			if _, ok := ctx.Value(OAuth2ClientDetailsKey).(model.ClientDetails); !ok {
				return nil, ErrInvalidClientRequest
			}
			return next(ctx, request)
		}
	}
}

// 创建验权中间件
func MakeAuthorityAuthorizationMiddleware(authority string, logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if err, ok := ctx.Value(OAuth2ErrorKey).(error); ok {
				return nil, err
			}
			if details, ok := ctx.Value(OAuth2DetailsKey).(model.OAuth2Details); !ok {
				return nil, ErrInvalidClientRequest
			} else {
				for _, value := range details.User.Authorities {
					if value == authority {
						return next(ctx, request)
					}
				}
				return nil, ErrNotPermit
			}
		}
	}
}

//  创建令牌终端
func MakeTokenEndpoint(svc service.TokenGranter, clientService service.ClientDetailsService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*TokenRequest)
		token, err := svc.Grant(ctx, req.GrantType, ctx.Value(OAuth2ClientDetailsKey).(model.ClientDetails), req.Reader)
		var errString = ""
		if err != nil {
			errString = err.Error()
		}

		return TokenResponse{
			AccessToken: token,
			Error:       errString,
		}, nil
	}
}

// 创建校验令牌终端
func MakeCheckTokenEndpoint(svc service.TokenService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(*CheckTokenRequest)
		tokenDetails, err := svc.GetOAuth2DetailsByAccessToken(req.Token)
		var errString = ""
		if err != nil {
			errString = err.Error()
		}
		return CheckTokenResponse{
			OAuthDetails: tokenDetails,
			Error:        errString,
		}, nil
	}
}
