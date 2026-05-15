package memberapi

import (
	"encoding/json"
	"net/http"

	"parent-child-api/api/apix"
	"parent-child-api/model"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type MemberApi struct{}

func (x MemberApi) CreateVirtualChild(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.VirtualChildCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	apix.WriteData(w, srv.MemberService.CreateVirtualChild(token.UserId, req))
}
