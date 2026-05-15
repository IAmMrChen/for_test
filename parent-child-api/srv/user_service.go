package srv

import (
	"parent-child-api/model"
	"parent-child-api/resx"
	"parent-child-api/ux"
)

var UserService userService

type userService struct{}

func (x userService) GetUser(userId int64) *model.User {
	const sql = `
		SELECT id
			, openid
			, IFNULL(nickname, '') AS nickname
			, IFNULL(avatar_url, '') AS avatar_url
		FROM users
		WHERE id=@p1
	`

	user := &model.User{}
	ok := resx.Db.Main.MustGetStruct(user, sql, userId)
	if !ok {
		return nil
	}
	return user
}

func (x userService) LoadCurrentUser(token *ux.AuthTokenClaims) model.User {
	user := x.GetUser(token.UserId)
	if user != nil {
		return *user
	}

	return model.User{
		Id:       token.UserId,
		Openid:   token.OpenId,
		Nickname: token.Nickname,
	}
}
