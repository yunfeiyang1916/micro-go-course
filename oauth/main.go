package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/yunfeiyang1916/micro-go-course/oauth/config"
	"github.com/yunfeiyang1916/micro-go-course/oauth/endpoint"
	"github.com/yunfeiyang1916/micro-go-course/oauth/model"
	"github.com/yunfeiyang1916/micro-go-course/oauth/service"
	"github.com/yunfeiyang1916/micro-go-course/oauth/transport"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	var (
		servicePort = flag.Int("service.port", 10086, "service port")
	)

	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	// 令牌服务
	var tokenService service.TokenService
	// 令牌生成器
	var tokenGranter service.TokenGranter
	// 令牌组装
	var tokenEnhancer service.TokenEnhancer
	// 令牌存储
	var tokenStore service.TokenStore
	// 用户详情服务
	var userDetailsService service.UserDetailsService
	// 客户端详情服务
	var clientDetailsService service.ClientDetailsService
	var srv service.Service

	tokenEnhancer = service.NewJwtTokenEnhancer("secret")
	tokenStore = service.NewJwtTokenStore(tokenEnhancer.(*service.JwtTokenEnhancer))
	tokenService = service.NewTokenService(tokenStore, tokenEnhancer)

	userDetailsService = service.NewInMemoryUserDetailsService([]*model.UserDetails{
		{
			Username:    "aoho",
			Password:    "123456",
			UserId:      1,
			Authorities: []string{"Simple"},
		},
		{
			Username:    "admin",
			Password:    "123456",
			UserId:      1,
			Authorities: []string{"Admin"},
		},
	})

	clientDetailsService = service.NewInMemoryClientDetailService([]*model.ClientDetails{
		{
			ClientId:                    "clientId",
			ClientSecret:                "clientSecret",
			AccessTokenValiditySeconds:  1800,
			RefreshTokenValiditySeconds: 18000,
			RegisteredRedirectUri:       "http://127.0.0.1",
			AuthorizedGrantTypes:        []string{"password", "refresh_token"},
		},
	})
	tokenGranter = service.NewComposeTokenGranter(map[string]service.TokenGranter{
		"password":      service.NewUsernamePasswordTokenGranter("password", userDetailsService, tokenService),
		"refresh_token": service.NewRefreshGranter("refresh_token", userDetailsService, tokenService),
	})
	tokenEndpoint := endpoint.MakeTokenEndpoint(tokenGranter, clientDetailsService)
	tokenEndpoint = endpoint.MakeClientAuthorizationMiddleware(config.KitLogger)(tokenEndpoint)
	checkTokenEndpoint := endpoint.MakeCheckTokenEndpoint(tokenService)
	checkTokenEndpoint = endpoint.MakeClientAuthorizationMiddleware(config.KitLogger)(checkTokenEndpoint)

	srv = service.NewCommonService()
	//创建健康检查的Endpoint
	healthEndpoint := endpoint.MakeHealthCheckEndpoint(srv)

	endpts := endpoint.OAuth2Endpoints{
		TokenEndpoint:       tokenEndpoint,
		CheckTokenEndpoint:  checkTokenEndpoint,
		HealthCheckEndpoint: healthEndpoint,
	}

	// 创建http.Handler
	r := transport.MakeHttpHandler(ctx, endpts, tokenService, clientDetailsService, config.KitLogger)
	go func() {
		config.Logger.Println("Http Server start at port:" + strconv.Itoa(*servicePort))
		handler := r
		errChan <- http.ListenAndServe(":"+strconv.Itoa(*servicePort), handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan
	config.Logger.Println(error)
}
