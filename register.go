package sip

type RegisterInfo struct {
	Registrar Connectinfo
	Client    Connectinfo
	Username  string
	UserInfo  UserInfo

	Expiration int
}
