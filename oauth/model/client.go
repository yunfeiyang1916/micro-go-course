package model

// 客户端详情
type ClientDetails struct {
	// 客户端标识
	ClientId string
	// 客户端的密钥
	ClientSecret string
	// 访问令牌有效时间，秒
	AccessTokenValiditySeconds int
	// 刷新令牌有效时间，秒
	RefreshTokenValiditySeconds int
	// 重定向地址，授权码类型中使用
	RegisteredRedirectUri string
	// 可以使用的授权类型
	AuthorizedGrantTypes []string
}
