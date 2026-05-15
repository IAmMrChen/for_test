package inviteapi

import (
	"encoding/json"
	"net/http"

	"parent-child-api/api/apix"
	"parent-child-api/model"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type InviteApi struct{}

func (x InviteApi) Create(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.FamilyInviteCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	apix.WriteData(w, srv.InviteService.CreateInvite(token.UserId, req))
}

func (x InviteApi) Accept(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.FamilyInviteAcceptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	apix.WriteData(w, srv.InviteService.AcceptInvite(token.UserId, req.Token))
}
