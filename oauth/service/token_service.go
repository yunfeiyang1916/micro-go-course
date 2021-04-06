package service

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"

	"github.com/yunfeiyang1916/micro-go-course/oauth/model"
)

var (
	// 授权类型错误
	ErrNotSupportGrantType = errors.New("grant type is not supported")
	// 不支持的操作
	ErrNotSupportOperation = errors.New("no support operation")
	// 无效的用户名和密码
	ErrInvalidUsernameAndPasswordRequest = errors.New("invalid username, password")
	ErrInvalidTokenRequest               = errors.New("invalid token")
	// 令牌过期
	ErrExpiredToken = errors.New("token is expired")
)

// 令牌生成器
type TokenGranter interface {
	// 生成令牌
	Grant(ctx context.Context, grantType string, client model.ClientDetails, reader *http.Request) (*model.OAuth2Token, error)
}

// 令牌服务接口
type TokenService interface {
	// 根据访问令牌获取对应的用户信息和客户端信息
	GetOAuth2DetailsByAccessToken(tokenValue string) (*model.OAuth2Details, error)
	// 生成访问令牌
	CreateAccessToken(oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error)
	// 根据刷新令牌获取访问令牌
	RefreshAccessToken(refreshTokenValue string) (*model.OAuth2Token, error)
	// 根据用户信息和客户端信息获取已生成访问令牌
	GetAccessToken(details *model.OAuth2Details) (*model.OAuth2Token, error)
	// 根据访问令牌值获取访问令牌结构体
	ReadAccessToken(tokenValue string) (*model.OAuth2Token, error)
}

// 令牌存储
type TokenStore interface {
	// 存储访问令牌
	StoreAccessToken(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details) error
	// 根据令牌值获取访问令牌结构体
	ReadAccessToken(tokenValue string) (*model.OAuth2Token, error)
	// 根据令牌值获取令牌对应的客户端和用户信息
	ReadOAuth2Details(tokenValue string) (*model.OAuth2Details, error)
	// 根据客户端信息和用户信息获取访问令牌
	GetAccessToken(oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error)
	// 移除存储的访问令牌
	RemoveAccessToken(tokenValue string) error
	// 存储刷新令牌
	StoreRefreshToken(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details) error
	// 移除存储的刷新令牌
	RemoveRefreshToken(oauth2Token string) error
	// 根据令牌值获取刷新令牌
	ReadRefreshToken(tokenValue string) (*model.OAuth2Token, error)
	// 根据令牌值获取刷新令牌对应的客户端和用户信息
	ReadOAuth2DetailsForRefreshToken(tokenValue string) (*model.OAuth2Details, error)
}

// 组合模式令牌生成器，管理了多种 LeafTokenGranter 授权类型的具体叶节点实现
type ComposeTokenGranter struct {
	// 令牌生成器字典,以授权类型为键，授权生成器为值
	TokenGrantDict map[string]TokenGranter
}

func NewComposeTokenGranter(tokenGrantDict map[string]TokenGranter) TokenGranter {
	return &ComposeTokenGranter{
		TokenGrantDict: tokenGrantDict,
	}
}

// 生成令牌
func (c *ComposeTokenGranter) Grant(ctx context.Context, grantType string, client model.ClientDetails, reader *http.Request) (*model.OAuth2Token, error) {
	// 检查客户端是否允许该种授权类型
	var isSupport bool
	if len(client.AuthorizedGrantTypes) > 0 {
		for _, v := range client.AuthorizedGrantTypes {
			if v == grantType {
				isSupport = true
				break
			}
		}
	}
	if !isSupport {
		return nil, ErrNotSupportOperation
	}
	// 查找具体的授权类型实现节点
	if dispatchGranter, ok := c.TokenGrantDict[grantType]; ok {
		return dispatchGranter.Grant(ctx, grantType, client, reader)
	}
	return nil, ErrNotSupportGrantType
}

// 使用用户名与密码令牌生成器
type UsernamePasswordTokenGranter struct {
	// 支持的授权类型
	supportGrantType string
	// 用户详情服务
	userDetailsService UserDetailsService
	// 令牌服务
	tokenService TokenService
}

// 生成令牌
func (u *UsernamePasswordTokenGranter) Grant(ctx context.Context, grantType string, client model.ClientDetails, reader *http.Request) (*model.OAuth2Token, error) {
	if grantType != u.supportGrantType {
		return nil, ErrNotSupportGrantType
	}
	// 从请求体获取用户名和密码
	username := reader.FormValue("username")
	password := reader.FormValue("password")
	if username == "" || password == "" {
		return nil, ErrInvalidUsernameAndPasswordRequest
	}
	// 验证用户名密码是否正确
	userDetails, err := u.userDetailsService.GetUserDetailByUsername(ctx, username, password)
	if err != nil {
		return nil, ErrInvalidUsernameAndPasswordRequest
	}
	// 根据用户信息和客户端信息生成访问令牌
	return u.tokenService.CreateAccessToken(&model.OAuth2Details{
		User:   userDetails,
		Client: client,
	})
}

func NewUsernamePasswordTokenGranter(grantType string, userDetailsService UserDetailsService, tokenService TokenService) TokenGranter {
	return &UsernamePasswordTokenGranter{
		supportGrantType:   grantType,
		userDetailsService: userDetailsService,
		tokenService:       tokenService,
	}
}

// 刷新令牌生成器
type RefreshTokenGranter struct {
	// 支持的授权类型
	supportGrantType string
	// 令牌服务
	tokenService TokenService
}

// 生成令牌
func (r *RefreshTokenGranter) Grant(ctx context.Context, grantType string, client model.ClientDetails, reader *http.Request) (*model.OAuth2Token, error) {
	if grantType != r.supportGrantType {
		return nil, ErrNotSupportGrantType
	}
	// 从请求体获取刷新令牌
	refreshTokenValue := reader.URL.Query().Get("refresh_token")
	if refreshTokenValue == "" {
		return nil, ErrInvalidTokenRequest
	}
	return r.tokenService.RefreshAccessToken(refreshTokenValue)
}
func NewRefreshGranter(grantType string, userDetailsService UserDetailsService, tokenService TokenService) TokenGranter {
	return &RefreshTokenGranter{
		supportGrantType: grantType,
		tokenService:     tokenService,
	}
}

// 令牌服务默认实现
type DefaultTokenService struct {
	// 令牌存储
	tokenStore TokenStore
	// 令牌组装
	tokenEnhancer TokenEnhancer
}

func NewTokenService(tokenStore TokenStore, tokenEnhancer TokenEnhancer) TokenService {
	return &DefaultTokenService{
		tokenStore:    tokenStore,
		tokenEnhancer: tokenEnhancer,
	}
}

// 根据访问令牌获取对应的用户信息和客户端信息
func (d *DefaultTokenService) GetOAuth2DetailsByAccessToken(tokenValue string) (*model.OAuth2Details, error) {
	accessToken, err := d.tokenStore.ReadAccessToken(tokenValue)
	if err != nil {
		return nil, err
	}
	if accessToken.IsExpired() {
		return nil, ErrExpiredToken
	}
	return d.tokenStore.ReadOAuth2Details(tokenValue)
}

// 生成访问令牌
func (d *DefaultTokenService) CreateAccessToken(oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error) {
	existToken, err := d.tokenStore.GetAccessToken(oauth2Details)
	if err != nil {
		return nil, err
	}
	var refreshToken *model.OAuth2Token
	// 存在未失效的访问令牌，直接返回
	if existToken != nil {
		if !existToken.IsExpired() {
			return existToken, nil
		}
		// 已过期，移除令牌
		err = d.tokenStore.RemoveAccessToken(existToken.TokenValue)
		if err != nil {
			return nil, err
		}
		// 移除刷新令牌
		if existToken.RefreshToken != nil {
			refreshToken = existToken.RefreshToken
			err = d.tokenStore.RemoveRefreshToken(refreshToken.TokenType)
			if err != nil {
				return nil, err
			}
		}
	}
	if refreshToken == nil || refreshToken.IsExpired() {
		refreshToken, err = d.createRefreshToken(oauth2Details)
		if err != nil {
			return nil, err
		}
	}
	// 生成新的访问令牌
	accessToken, err := d.createAccessToken(refreshToken, oauth2Details)
	if err != nil {
		return nil, err
	}
	// 保存新生成令牌
	err = d.tokenStore.StoreAccessToken(accessToken, oauth2Details)
	if err != nil {
		return nil, err
	}
	err = d.tokenStore.StoreRefreshToken(refreshToken, oauth2Details)
	if err != nil {
		return nil, err
	}
	return accessToken, err
}

// 创建访问令牌
func (d *DefaultTokenService) createAccessToken(refreshToken *model.OAuth2Token, oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error) {
	validitySeconds := oauth2Details.Client.AccessTokenValiditySeconds
	s, _ := time.ParseDuration(strconv.Itoa(validitySeconds) + "s")
	expiredTime := time.Now().Add(s)
	accessToken := &model.OAuth2Token{
		RefreshToken: refreshToken,
		ExpiresTime:  &expiredTime,
		TokenValue:   uuid.NewV4().String(),
	}
	if d.tokenEnhancer != nil {
		return d.tokenEnhancer.Enhance(accessToken, oauth2Details)
	}
	return refreshToken, nil
}

// 创建刷新令牌
func (d *DefaultTokenService) createRefreshToken(oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error) {
	validitySeconds := oauth2Details.Client.RefreshTokenValiditySeconds
	s, _ := time.ParseDuration(strconv.Itoa(validitySeconds) + "s")
	expiredTime := time.Now().Add(s)
	refreshToken := &model.OAuth2Token{
		ExpiresTime: &expiredTime,
		TokenValue:  uuid.NewV4().String(),
	}

	if d.tokenEnhancer != nil {
		return d.tokenEnhancer.Enhance(refreshToken, oauth2Details)
	}
	return refreshToken, nil
}

// 根据刷新令牌获取访问令牌
func (d *DefaultTokenService) RefreshAccessToken(refreshTokenValue string) (*model.OAuth2Token, error) {
	refreshToken, err := d.tokenStore.ReadRefreshToken(refreshTokenValue)
	if err == nil {
		if refreshToken.IsExpired() {
			return nil, ErrExpiredToken
		}
		oauth2Details, err := d.tokenStore.ReadOAuth2DetailsForRefreshToken(refreshTokenValue)
		if err == nil {
			oauth2Token, err := d.tokenStore.GetAccessToken(oauth2Details)
			// 移除原有的访问令牌
			if err == nil {
				d.tokenStore.RemoveAccessToken(oauth2Token.TokenValue)
			}

			// 移除已使用的刷新令牌
			d.tokenStore.RemoveRefreshToken(refreshTokenValue)
			refreshToken, err = d.createRefreshToken(oauth2Details)
			if err == nil {
				accessToken, err := d.createAccessToken(refreshToken, oauth2Details)
				if err == nil {
					d.tokenStore.StoreAccessToken(accessToken, oauth2Details)
					d.tokenStore.StoreRefreshToken(refreshToken, oauth2Details)
				}
				return accessToken, err
			}
		}
	}
	return nil, err
}

// 根据用户信息和客户端信息获取已生成访问令牌
func (d *DefaultTokenService) GetAccessToken(details *model.OAuth2Details) (*model.OAuth2Token, error) {
	return d.tokenStore.GetAccessToken(details)
}

// 根据访问令牌值获取访问令牌结构体
func (d *DefaultTokenService) ReadAccessToken(tokenValue string) (*model.OAuth2Token, error) {
	return d.tokenStore.ReadAccessToken(tokenValue)
}

// jwt令牌存储
type JwtTokenStore struct {
	jwtTokenEnhancer *JwtTokenEnhancer
}

func NewJwtTokenStore(enhancer *JwtTokenEnhancer) TokenStore {
	return &JwtTokenStore{
		jwtTokenEnhancer: enhancer,
	}
}

// 存储访问令牌
func (j *JwtTokenStore) StoreAccessToken(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details) error {
	return nil
}

// 根据令牌值获取访问令牌结构体
func (j *JwtTokenStore) ReadAccessToken(tokenValue string) (*model.OAuth2Token, error) {
	oauth2Token, _, err := j.jwtTokenEnhancer.Extract(tokenValue)
	return oauth2Token, err
}

// 根据令牌值获取令牌对应的客户端和用户信息
func (j *JwtTokenStore) ReadOAuth2Details(tokenValue string) (*model.OAuth2Details, error) {
	_, oauth2Details, err := j.jwtTokenEnhancer.Extract(tokenValue)
	return oauth2Details, err
}

// 根据客户端信息和用户信息获取访问令牌
func (j *JwtTokenStore) GetAccessToken(oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error) {
	return nil, nil
}

// 移除存储的访问令牌
func (j *JwtTokenStore) RemoveAccessToken(tokenValue string) error {
	return nil
}

// 存储刷新令牌
func (j *JwtTokenStore) StoreRefreshToken(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details) error {
	return nil
}

// 移除存储的刷新令牌
func (j *JwtTokenStore) RemoveRefreshToken(oauth2Token string) error {
	return nil
}

// 根据令牌值获取刷新令牌
func (j *JwtTokenStore) ReadRefreshToken(tokenValue string) (*model.OAuth2Token, error) {
	oauth2Token, _, err := j.jwtTokenEnhancer.Extract(tokenValue)
	return oauth2Token, err
}

// 根据令牌值获取刷新令牌对应的客户端和用户信息
func (j *JwtTokenStore) ReadOAuth2DetailsForRefreshToken(tokenValue string) (*model.OAuth2Details, error) {
	_, oauth2Details, err := j.jwtTokenEnhancer.Extract(tokenValue)
	return oauth2Details, err
}

// 令牌组装者接口
type TokenEnhancer interface {
	// 组装Token信息
	Enhance(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error)
	// 从Token中还原信息
	Extract(tokenValue string) (*model.OAuth2Token, *model.OAuth2Details, error)
}

// 令牌定制声明
type OAuth2TokenCustomClaims struct {
	// 用户详情
	UserDetails model.UserDetails
	// 客户端详情
	ClientDetails model.ClientDetails
	// 重新刷新令牌
	RefreshToken model.OAuth2Token
	jwt.StandardClaims
}

// jwt令牌组装者
type JwtTokenEnhancer struct {
	// 密钥
	secretKey []byte
}

// 组装Token信息
func (j *JwtTokenEnhancer) Enhance(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error) {
	return j.sign(oauth2Token, oauth2Details)
}

// 从Token中还原信息
func (j *JwtTokenEnhancer) Extract(tokenValue string) (*model.OAuth2Token, *model.OAuth2Details, error) {
	token, err := jwt.ParseWithClaims(tokenValue, &OAuth2TokenCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})
	if err != nil {
		return nil, nil, err
	}
	claims := token.Claims.(*OAuth2TokenCustomClaims)
	expiresTime := time.Unix(claims.ExpiresAt, 0)
	return &model.OAuth2Token{
			RefreshToken: &claims.RefreshToken,
			TokenValue:   tokenValue,
			ExpiresTime:  &expiresTime,
		}, &model.OAuth2Details{
			User:   claims.UserDetails,
			Client: claims.ClientDetails,
		},
		nil
}

// 签名
func (j *JwtTokenEnhancer) sign(oauth2Token *model.OAuth2Token, oauth2Details *model.OAuth2Details) (*model.OAuth2Token, error) {
	// 过期时间
	expireTime := oauth2Token.ExpiresTime
	clientDetails := oauth2Details.Client
	userDetails := oauth2Details.User
	clientDetails.ClientSecret = ""
	userDetails.Password = ""
	claims := OAuth2TokenCustomClaims{
		UserDetails:   userDetails,
		ClientDetails: clientDetails,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    "System",
		},
	}
	if oauth2Token.RefreshToken != nil {
		claims.RefreshToken = *oauth2Token.RefreshToken
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenValue, err := token.SignedString(j.secretKey)
	if err != nil {
		return nil, err
	}
	oauth2Token.TokenValue = tokenValue
	oauth2Token.TokenType = "jwt"
	return oauth2Token, nil
}

func NewJwtTokenEnhancer(secretKey string) TokenEnhancer {
	return &JwtTokenEnhancer{
		secretKey: []byte(secretKey),
	}
}
