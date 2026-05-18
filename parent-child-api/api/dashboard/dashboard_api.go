package dashboardapi

import (
	"net/http"
	"strconv"

	"parent-child-api/api/apix"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type DashboardApi struct{}

func (x DashboardApi) Summary(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}

	apix.WriteData(w, srv.DashboardService.Summary(token.UserId, familyId))
}
