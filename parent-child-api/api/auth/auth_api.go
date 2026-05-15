package authapi

import (
	"encoding/json"
	"net/http"

	"parent-child-api/api/apix"
	"parent-child-api/model"
	"parent-child-api/srv"
)

type AuthApi struct {
	JwtSecret string
}

func (x AuthApi) DemoLogin(w http.ResponseWriter, r *http.Request) {
	var req model.DemoLoginRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	res := srv.AuthService.CreateDemoToken(req, x.JwtSecret)
	apix.WriteData(w, res)
}
