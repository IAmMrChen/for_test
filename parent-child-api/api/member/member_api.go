package memberapi

import (
	"encoding/json"
	"net/http"
	"strconv"

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

func (x MemberApi) List(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}

	apix.WriteData(w, srv.MemberService.ListMembers(token.UserId, familyId))
}
