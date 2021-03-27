package service

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
	"log"
	"micro-go-course/user-server/dao"
	"micro-go-course/user-server/redis"
	"time"
)

// 用户信息数据传输对象
type UserInfoDTO struct {
	ID int64 `json:"id"`
	// 用户名
	Username string `json:"username"`
	// 邮箱
	Email string `json:"email"`
}

// 用户注册视图对象
type RegisterUserVO struct {
	Username string
	Password string
	Email    string
}

var (
	ErrUserExisted = errors.New("user is existed")
	ErrPassword    = errors.New("email and password are not match")
	ErrRegistering = errors.New("email is registering")
)

// 用户服务接口
type UserService interface {
	// 登录
	Login(ctx context.Context, email, password string) (*UserInfoDTO, error)
	// 注册
	Register(ctx context.Context, vo *RegisterUserVO) (*UserInfoDTO, error)
}

// 用户服务实现
type UserServiceImpl struct {
	userDAO dao.UserDAO
}

func MakeUserServiceImpl(userDAO dao.UserDAO) UserService {
	return &UserServiceImpl{
		userDAO: userDAO,
	}
}

// 登录
func (u *UserServiceImpl) Login(ctx context.Context, email, password string) (*UserInfoDTO, error) {
	user, err := u.userDAO.SelectByEmail(email)
	if err != nil {
		return nil, err
	}
	if user != nil && user.Password == password {
		return &UserInfoDTO{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}, nil
	}
	return nil, err
}

// 注册
func (u *UserServiceImpl) Register(ctx context.Context, vo *RegisterUserVO) (*UserInfoDTO, error) {
	lock := redis.GetRedisLock(vo.Email, 5*time.Second)
	// 加分布式锁
	err := lock.Lock()
	if err != nil {
		log.Printf("err:%s", err)
		return nil, ErrRegistering
	}
	defer lock.Unlock()
	existUser, err := u.userDAO.SelectByEmail(vo.Email)
	if (err == nil && existUser == nil) || err == gorm.ErrRecordNotFound {
		newUser := &dao.UserEntity{
			Username: vo.Username,
			Password: vo.Password,
			Email:    vo.Email,
		}
		err = u.userDAO.Save(newUser)
		if err == nil {
			return &UserInfoDTO{
				ID:       newUser.ID,
				Username: newUser.Username,
				Email:    newUser.Email,
			}, nil
		}
	}
	if err == nil {
		err = ErrUserExisted
	}
	return nil, err
}
