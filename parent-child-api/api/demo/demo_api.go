package demoapi

import (
	"net/http"

	"parent-child-api/api/apix"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type DemoApi struct{}

func (x DemoApi) Me(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	apix.WriteData(w, srv.UserService.LoadCurrentUser(token))
}
