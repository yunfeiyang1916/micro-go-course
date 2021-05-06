package service

import "fmt"

type Service interface {
	// HealthCheck check service health status
	HealthCheck() bool
	// 首页，任意请求可访问
	Index() string
	// 示例，携带有效访问令牌的请求可访问
	Sample(username string) string
	// 携带有效访问令牌，且访问令牌绑定的用户具备 Admin 权限的请求可访问
	Admin(username string) string
}

type CommonService struct {
}

// HealthCheck implement Service method
// 用于检查服务的健康状态，这里仅仅返回true
func (s *CommonService) HealthCheck() bool {
	return true
}

// 首页，任意请求可访问
func (s *CommonService) Index() string {
	return fmt.Sprintf("hello, wecome to index")
}

// 示例，携带有效访问令牌的请求可访问
func (s *CommonService) Sample(username string) string {
	return fmt.Sprintf("hello %s, wecome to sample", username)
}

// 携带有效访问令牌，且访问令牌绑定的用户具备 Admin 权限的请求可访问
func (s *CommonService) Admin(username string) string {
	return fmt.Sprintf("hello %s, wecome to admin", username)

}

func NewCommonService() *CommonService {
	return &CommonService{}
}
