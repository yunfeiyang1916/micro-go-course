package user

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
)

// 中间件
type ServiceMiddleware func(service UserService) UserService

// 日志中间件
type loggingMiddleware struct {
	UserService
	logger log.Logger
}

// 日志中间件
func LoggingMiddleware(logger log.Logger) ServiceMiddleware {
	return func(next UserService) UserService {
		return loggingMiddleware{logger: logger}
	}
}

func (mw loggingMiddleware) CheckPassword(ctx context.Context, username, password string) (ret bool, err error) {
	// 记录日志
	defer func(begin time.Time) {
		mw.logger.Log("func", "CheckPassword", "username", username, "password", password, "result", ret, "took", time.Since(begin))
	}(time.Now())
	ret, err = mw.UserService.CheckPassword(ctx, username, password)
	return
}
