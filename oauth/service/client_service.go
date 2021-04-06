package service

import (
	"context"
	"errors"

	"github.com/yunfeiyang1916/micro-go-course/oauth/model"
)

var (
	// 客户端标识不存在
	ErrClientNotExist = errors.New("clientId is not exist")
	// 客户端密钥错误
	ErrClientSecret = errors.New("invalid clientSecret")
)

// 客户端详情服务接口
type ClientDetailsService interface {
	// 根据 clientId和密钥获取客户端信息
	GetClientDetailsByClientId(ctx context.Context, clientId, clientSecret string) (model.ClientDetails, error)
}

// 客户端详情服务实现
type InMemoryClientDetailsService struct {
	// 以客户端id为键，客户端详情为值的字典
	clientDetailsDict map[string]*model.ClientDetails
}

// 构造客户端详情服务实现实例
func NewInMemoryClientDetailService(clientDetailsList []*model.ClientDetails) *InMemoryClientDetailsService {
	clientDetailsDict := make(map[string]*model.ClientDetails)
	if len(clientDetailsList) > 0 {
		for _, value := range clientDetailsList {
			clientDetailsDict[value.ClientId] = value
		}
	}
	return &InMemoryClientDetailsService{
		clientDetailsDict: clientDetailsDict,
	}
}

// 根据 clientId和密钥获取客户端信息
func (service *InMemoryClientDetailsService) GetClientDetailsByClientId(ctx context.Context, clientId, clientSecret string) (model.ClientDetails, error) {
	clientDetails, ok := service.clientDetailsDict[clientId]
	if !ok {
		return model.ClientDetails{}, ErrClientNotExist
	}
	// 密码是否正确
	if clientDetails.ClientSecret != clientSecret {
		return model.ClientDetails{}, ErrClientSecret
	}
	return *clientDetails, nil
}
