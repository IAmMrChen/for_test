package taskapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"parent-child-api/api/apix"
	"parent-child-api/model"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type TaskApi struct{}

func (x TaskApi) Create(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.TaskCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.TaskService.CreateTask(token.UserId, req))
}

func (x TaskApi) Update(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.TaskUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.TaskService.UpdateTask(token.UserId, req))
}

func (x TaskApi) Archive(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.TaskArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.TaskService.ArchiveTask(token.UserId, req))
}

func (x TaskApi) List(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}
	apix.WriteData(w, srv.TaskService.ListTasks(token.UserId, familyId))
}

func (x TaskApi) Claim(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.TaskClaimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.TaskService.ClaimTask(token.UserId, req))
}

func (x TaskApi) Submit(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.TaskSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.TaskService.SubmitTask(token.UserId, req))
}

func (x TaskApi) Audit(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.TaskAuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.TaskService.AuditTask(token.UserId, req))
}

func (x TaskApi) Records(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}

	req := model.TaskRecordListRequest{
		FamilyId: familyId,
		Status:   model.TaskRecordStatus(r.URL.Query().Get("status")),
	}
	apix.WriteData(w, srv.TaskService.ListTaskRecords(token.UserId, req))
}
