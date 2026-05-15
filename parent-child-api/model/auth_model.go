package model

type DemoLoginRequest struct {
	UserId   int64  `json:"userId"`
	OpenId   string `json:"openid"`
	Nickname string `json:"nickname"`
}

type DemoLoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
