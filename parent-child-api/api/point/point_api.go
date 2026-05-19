package pointapi

import (
	"net/http"
	"strconv"

	"parent-child-api/api/apix"
	"parent-child-api/model"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type PointApi struct{}

func (x PointApi) Logs(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}

	memberId, err := parseOptionalInt64(r.URL.Query().Get("memberId"))
	if err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid memberId")
		return
	}

	apix.WriteData(w, srv.PointService.ListPointLogs(token.UserId, model.PointLogListRequest{
		FamilyId: familyId,
		MemberId: memberId,
	}))
}

func parseOptionalInt64(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}
