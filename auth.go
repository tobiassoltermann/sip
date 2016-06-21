package sip

type UserInfo interface {
	GetUsername() string
	GetType() string
}

type UserInfoImpl struct {
	username string
}

func (u *UserInfoImpl) GetUsername() string {
	return u.username
}
func (u *UserInfoImpl) GetType() string {
	return "UNAUTHORIZED"
}

type DigestUserInfoImpl struct {
	UserInfoImpl
	password string
}

func (d *DigestUserInfoImpl) GetPassword() string {
	return d.password
}
func (d *DigestUserInfoImpl) GetType() string {
	return "DIGEST"
}

func UnauthorizedUserInfo(username string) UserInfo {
	return &UserInfoImpl{
		username,
	}
}

func DigestUserInfo(username string, password string) UserInfo {
	return &DigestUserInfoImpl{
		UserInfoImpl{
			username,
		},
		password,
	}
}
