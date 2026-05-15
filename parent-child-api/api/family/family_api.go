package familyapi

import (
	"encoding/json"
	"net/http"

	"parent-child-api/api/apix"
	"parent-child-api/model"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type FamilyApi struct{}

func (x FamilyApi) Create(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.FamilyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	apix.WriteData(w, srv.FamilyService.CreateFamily(token.UserId, req))
}

func (x FamilyApi) List(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	apix.WriteData(w, srv.FamilyService.ListFamilies(token.UserId))
}
