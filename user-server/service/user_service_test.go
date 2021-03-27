package service

import (
	"context"
	"github.com/yunfeiyang1916/micro-go-course/user-server/dao"
	"github.com/yunfeiyang1916/micro-go-course/user-server/redis"
	"testing"
)

func TestUserServiceImpl_Login(t *testing.T) {
	err := dao.InitMysql("127.0.0.1", "3306", "root", "root123456", "user")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = redis.InitRedis("127.0.0.1", "6379", "")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	userService := MakeUserServiceImpl(&dao.UserDAOImpl{})
	user, err := userService.Login(context.TODO(), "zhangsan@163.com", "abcde")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("user id is %d", user.ID)
}

func TestUserServiceImpl_Register(t *testing.T) {
	err := dao.InitMysql("127.0.0.1", "3306", "root", "root123456", "user")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = redis.InitRedis("127.0.0.1", "6379", "")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	userService := MakeUserServiceImpl(&dao.UserDAOImpl{})
	user, err := userService.Register(context.TODO(), &RegisterUserVO{Username: "李四", Email: "lisi@163.com", Password: "lisi"})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("user id is %d", user.ID)
}
