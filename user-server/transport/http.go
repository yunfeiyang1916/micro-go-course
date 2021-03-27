package transport

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"micro-go-course/user-server/endpoint"
	"net/http"
	"os"
)

// 传输层，对外暴露项目的服务接口

var (
	ErrorBadRequest = errors.New("invalid request parameter")
)

// http处理请求
func MakeHttpHandler(ctx context.Context, endpoints *endpoint.UserEndpoints) http.Handler {
	r := mux.NewRouter()
	kitLog := log.NewLogfmtLogger(os.Stderr)
	kitLog = log.With(kitLog, "ts", log.DefaultTimestampUTC)
	kitLog = log.With(kitLog, "caller", log.DefaultCaller)

	options := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(kitLog)),
		kithttp.ServerErrorEncoder(encodeError),
	}
	// 用户注册路由
	r.Methods("POST").Path("/register").Handler(kithttp.NewServer(endpoints.RegisterEndpoint, decodeRegisterRequest, encodeJSONResponse, options...))
	// 用户登录路由
	r.Methods("POST").Path("/login").Handler(kithttp.NewServer(endpoints.LoginEndpoint, decodeLoginRequest, encodeJSONResponse, options...))

	return r
}

// 用户注册请求解码
func decodeRegisterRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")

	if username == "" || password == "" || email == "" {
		return nil, ErrorBadRequest
	}
	return &endpoint.RegisterRequest{
		Username: username,
		Password: password,
		Email:    email,
	}, nil
}

// 用户登录请求解码
func decodeLoginRequest(_ context.Context, r *http.Request) (interface{}, error) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		return nil, ErrorBadRequest
	}
	return &endpoint.LoginReq{
		Email:    email,
		Password: password,
	}, nil
}

// json输出
func encodeJSONResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// 错误编码处理
func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
