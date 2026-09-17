package logic

import (
	"errors"
	"strings"
	"testing"
)

func TestUserPasswordValid(t *testing.T) {
	cases := []struct {
		name     string
		userName string
		password string
		wantErr  error
	}{
		{name: "正常", userName: "zhangsan", password: "123456", wantErr: nil},
		{name: "用户名过短", userName: "a", password: "123456", wantErr: ErrorUserNameInvalid},
		{name: "用户名过长", userName: strings.Repeat("a", UserNameMaxLen+1), password: "123456", wantErr: ErrorUserNameInvalid},
		{name: "密码过短", userName: "zhangsan", password: "12345", wantErr: ErrorPasswordInvalid},
		{name: "密码过长", userName: "zhangsan", password: strings.Repeat("a", PasswordMaxLen+1), wantErr: ErrorPasswordInvalid},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := UserPasswordValid(c.userName, c.password)
			if !errors.Is(err, c.wantErr) {
				t.Errorf("UserPasswordValid(%q, %q) = %v, want %v", c.userName, c.password, err, c.wantErr)
			}
		})
	}
}
