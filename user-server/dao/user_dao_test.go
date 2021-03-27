package dao

import "testing"

func TestUserDaoImpl_Save(t *testing.T) {
	userDao := &UserDAOImpl{}
	err := InitMysql("127.0.0.1", "3306", "root", "root123456", "user")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	user := &UserEntity{
		Username: "张三",
		Password: "abcde",
		Email:    "zhangsan@163.com",
	}
	err = userDao.Save(user)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("new User ID is %d", user.ID)
}

func TestUserDaoImpl_SelectByEmail(t *testing.T) {
	userDao := &UserDAOImpl{}
	err := InitMysql("127.0.0.1", "3306", "root", "root123456", "user")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	user, err := userDao.SelectByEmail("zhangsan@163.com")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("result username is %s", user.Username)
}
