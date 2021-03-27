package dao

import "time"

// 用户实体
type UserEntity struct {
	ID int64
	// 用户名
	Username string
	// 密码
	Password string
	// 邮箱
	Email string
	// 创建日期
	CreatedAt time.Time
}

// 表名
func (UserEntity) TableName() string {
	return "user"
}

// 用户数据访问接口
type UserDAO interface {
	// 根据邮箱查询
	SelectByEmail(email string) (*UserEntity, error)
	// 保存
	Save(user *UserEntity) error
}

// 用户数据访问实现
type UserDAOImpl struct {
}

// 根据邮箱查询
func (u *UserDAOImpl) SelectByEmail(email string) (*UserEntity, error) {
	user := &UserEntity{}
	err := db.Where("email=?", email).First(user).Error
	return user, err
}

// 保存
func (u *UserDAOImpl) Save(user *UserEntity) error {
	return db.Create(user).Error
}
