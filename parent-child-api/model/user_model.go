package model

type User struct {
	Id        int64  `json:"id"`
	Openid    string `json:"openid"`
	Nickname  string `json:"nickname"`
	AvatarUrl string `json:"avatarUrl"`
}
