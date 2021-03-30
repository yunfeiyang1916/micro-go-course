package service

import "errors"

const (
	StrMaxSize = 1024
)

// Service errors
var (
	ErrMaxSize = errors.New("maximum size of 1024 bytes exceeded")

	ErrStrValue = errors.New("maximum size of 1024 bytes exceeded")
)

// 服务接口
type Service interface {
	// 拼接字符串
	Concat(req StringReq, ret *string) error
}

// 实现
type StringService struct {
}

// 拼接字符串
func (s StringService) Concat(req StringReq, ret *string) error {
	if len(req.A)+len(req.B) > StrMaxSize {
		*ret = ""
		return ErrMaxSize
	}
	*ret = req.A + req.B
	return nil
}
