package srv

import (
	"time"

	"parent-child-api/model"
	"parent-child-api/ux"
)

var AuthService authService

type authService struct{}

func (x authService) CreateDemoToken(req model.DemoLoginRequest, jwtSecret string) model.DemoLoginResponse {
	user := model.User{
		Id:       req.UserId,
		Openid:   req.OpenId,
		Nickname: req.Nickname,
	}
	if user.Id == 0 {
		user.Id = 1
	}
	if user.Openid == "" {
		user.Openid = "demo-openid"
	}
	if user.Nickname == "" {
		user.Nickname = "演示用户"
	}

	token, err := ux.JwtUtil.SignAuthToken(ux.AuthTokenClaims{
		UserId:   user.Id,
		OpenId:   user.Openid,
		Nickname: user.Nickname,
	}, jwtSecret, 24*time.Hour)
	if err != nil {
		panic(err)
	}

	return model.DemoLoginResponse{
		Token: token,
		User:  user,
	}
}
