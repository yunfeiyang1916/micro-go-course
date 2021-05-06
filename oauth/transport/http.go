package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yunfeiyang1916/micro-go-course/oauth/endpoint"
	"github.com/yunfeiyang1916/micro-go-course/oauth/service"
)

// 传输层，对外暴露项目的服务接口
var (
	// 错误请求
	ErrorBadRequest = errors.New("invalid request parameter")
	// 错误的授权类型
	ErrorGrantTypeRequest = errors.New("invalid request grant type")
	// 错误的令牌
	ErrorTokenRequest = errors.New("invalid request token")
	// 无效的客户端请求
	ErrInvalidClientRequest = errors.New("invalid client message")
)

// 创建http处理器
func MakeHttpHandler(ctx context.Context, endpoints endpoint.OAuth2Endpoints, tokenService service.TokenService, clientService service.ClientDetailsService, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}
	r.Path("/metrics").Handler(promhttp.Handler())
	clientAuthorizationOptions := []kithttp.ServerOption{
		kithttp.ServerBefore(makeClientAuthorizationContext(clientService, logger)),
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}
	oauth2AuthorizationOptions := []kithttp.ServerOption{
		kithttp.ServerBefore(makeOAuth2AuthorizationContext(tokenService, logger)),
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}
	// 用于客户端携带用户凭证请求访问令牌
	r.Methods("POST").Path("/oauth/token").Handler(kithttp.NewServer(endpoints.TokenEndpoint, decodeTokenRequest, encodeJsonResponse, clientAuthorizationOptions...))

	// 用于验证访问令牌的有效性，返回访问令牌绑定的客户端和用户信息
	r.Methods("POST").Path("/oauth/check_token").Handler(kithttp.NewServer(endpoints.TokenEndpoint, decodeCheckTokenRequest, encodeJsonResponse, clientAuthorizationOptions...))

	// create health check handler
	r.Methods("GET").Path("/health").Handler(kithttp.NewServer(endpoints.HealthCheckEndpoint, decodeHealthCheckRequest, encodeJsonResponse, options...))
	r.Methods("Get").Path("/index").Handler(kithttp.NewServer(endpoints.SampleEndpoint, decodeIndexRequest, encodeJsonResponse, oauth2AuthorizationOptions...))
	r.Methods("Get").Path("/sample").Handler(kithttp.NewServer(endpoints.SampleEndpoint, decodeSampleRequest, encodeJsonResponse, oauth2AuthorizationOptions...))
	r.Methods("Get").Path("/admin").Handler(kithttp.NewServer(endpoints.SampleEndpoint, decodeAdminRequest, encodeJsonResponse, oauth2AuthorizationOptions...))
	return r
}
func makeOAuth2AuthorizationContext(tokenService service.TokenService, logger log.Logger) kithttp.RequestFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		// 获取访问令牌
		accessTokenValue := r.Header.Get("Authorization")
		var err error
		if accessTokenValue != "" {
			// 获取令牌对应的用户信息和客户端信息
			oauth2Details, err := tokenService.GetOAuth2DetailsByAccessToken(accessTokenValue)
			if err == nil {
				return context.WithValue(ctx, endpoint.OAuth2DetailsKey, oauth2Details)
			}
		} else {
			err = ErrorTokenRequest
		}
		return context.WithValue(ctx, endpoint.OAuth2ErrorKey, err)
	}
}

// 创建客户端认证上下文
func makeClientAuthorizationContext(clientDetailsService service.ClientDetailsService, logger log.Logger) kithttp.RequestFunc {
	return func(ctx context.Context, request *http.Request) context.Context {
		if clientId, clientSecret, ok := request.BasicAuth(); ok {
			clientDetails, err := clientDetailsService.GetClientDetailsByClientId(ctx, clientId, clientSecret)
			if err != nil {
				return context.WithValue(ctx, endpoint.OAuth2ErrorKey, ErrInvalidClientRequest)
			}
			return context.WithValue(ctx, endpoint.OAuth2ClientDetailsKey, clientDetails)
		}
		return context.WithValue(ctx, endpoint.OAuth2ErrorKey, ErrInvalidClientRequest)
	}
}
func decodeIndexRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return &endpoint.IndexRequest{}, nil

}

func decodeSampleRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return &endpoint.SampleRequest{}, nil
}

func decodeAdminRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return &endpoint.AdminRequest{}, nil
}

// 解码令牌请求
func decodeTokenRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	grantType := r.URL.Query().Get("grant_type")
	if grantType == "" {
		return nil, ErrorGrantTypeRequest
	}
	return &endpoint.TokenRequest{
		GrantType: grantType,
		Reader:    r,
	}, nil
}

// 解码验证令牌请求
func decodeCheckTokenRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	tokenValue := r.URL.Query().Get("token")
	if tokenValue == "" {
		return nil, ErrorTokenRequest
	}
	return &endpoint.CheckTokenRequest{
		Token: tokenValue,
	}, nil
}
func encodeJsonResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// decodeHealthCheckRequest decode request
func decodeHealthCheckRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return endpoint.HealthRequest{}, nil
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
