package service

import (
	"context"
	"errors"

	"github.com/yunfeiyang1916/micro-go-course/oauth/model"
)

var (
	ErrUserNotExist = errors.New("username is not exist")
	ErrPassword     = errors.New("invalid password")
)

// 用户详情服务接口
type UserDetailsService interface {
	// 根据用户名和密码获取用户详情
	GetUserDetailByUsername(ctx context.Context, username, password string) (model.UserDetails, error)
}

// 用户详情服务
type InMemoryUserDetailsService struct {
	userDetailsDict map[string]*model.UserDetails
}

// 根据用户名和密码获取用户详情
func (service *InMemoryUserDetailsService) GetUserDetailByUsername(ctx context.Context, username, password string) (model.UserDetails, error) {
	// 根据 username 获取用户信息
	userDetails, ok := service.userDetailsDict[username]
	if ok {
		// 比较 password 是否匹配
		if userDetails.Password == password {
			return *userDetails, nil
		} else {
			return model.UserDetails{}, ErrPassword
		}
	} else {
		return model.UserDetails{}, ErrUserNotExist
	}

}

func NewInMemoryUserDetailsService(userDetailsList []*model.UserDetails) *InMemoryUserDetailsService {
	userDetailsDict := make(map[string]*model.UserDetails)
	if userDetailsList != nil {
		for _, value := range userDetailsList {
			userDetailsDict[value.Username] = value
		}
	}

	return &InMemoryUserDetailsService{
		userDetailsDict: userDetailsDict,
	}
}
