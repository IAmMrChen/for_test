package rewardapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"parent-child-api/api/apix"
	"parent-child-api/model"
	"parent-child-api/srv"
	"parent-child-api/ux"
)

type RewardApi struct{}

func (x RewardApi) Create(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.RewardCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.RewardService.CreateReward(token.UserId, req))
}

func (x RewardApi) Update(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.RewardUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.RewardService.UpdateReward(token.UserId, req))
}

func (x RewardApi) OffShelf(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.RewardOffShelfRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.RewardService.OffShelfReward(token.UserId, req))
}

func (x RewardApi) List(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}
	apix.WriteData(w, srv.RewardService.ListRewards(token.UserId, familyId))
}

func (x RewardApi) Apply(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.RewardApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.RewardService.ApplyReward(token.UserId, req))
}

func (x RewardApi) Deliver(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.RewardRecordOperateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.RewardService.DeliverReward(token.UserId, req))
}

func (x RewardApi) Reject(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.RewardRecordOperateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.RewardService.RejectReward(token.UserId, req))
}

func (x RewardApi) Receive(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	var req model.RewardRecordOperateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apix.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	apix.WriteData(w, srv.RewardService.ReceiveReward(token.UserId, req))
}

func (x RewardApi) Records(w http.ResponseWriter, r *http.Request, token *ux.AuthTokenClaims) {
	familyId, err := strconv.ParseInt(r.URL.Query().Get("familyId"), 10, 64)
	if err != nil || familyId == 0 {
		apix.WriteError(w, http.StatusBadRequest, "invalid familyId")
		return
	}
	req := model.RewardRecordListRequest{
		FamilyId: familyId,
		Status:   model.RewardRecordStatus(r.URL.Query().Get("status")),
	}
	apix.WriteData(w, srv.RewardService.ListRewardRecords(token.UserId, req))
}
