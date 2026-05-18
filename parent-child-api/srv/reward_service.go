package srv

import (
	"fmt"
	"strings"

	"parent-child-api/model"
	"parent-child-api/resx"
)

var RewardService rewardService

type rewardService struct{}

func normalizeRewardCreateRequest(req model.RewardCreateRequest) model.RewardCreateRequest {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		panic(fmt.Errorf("reward name is required"))
	}
	if req.PointsCost <= 0 {
		req.PointsCost = 1
	}
	if req.Stock == 0 {
		req.Stock = -1
	}
	return req
}

func (x rewardService) CreateReward(operatorUserId int64, req model.RewardCreateRequest) model.Reward {
	req = normalizeRewardCreateRequest(req)
	operator := MemberService.RequireParentRole(operatorUserId, req.FamilyId)

	const sql = `
		INSERT INTO rewards(family_id, name, points_cost, stock, status, created_by)
		VALUES(@p1, @p2, @p3, @p4, @p5, @p6)
	`
	resx.Db.Main.MustExecute(sql, req.FamilyId, req.Name, req.PointsCost, req.Stock, model.RewardStatusActive, operator.Id)

	rewardIdValue, ok := resx.Db.Main.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || rewardIdValue == nil {
		panic(fmt.Errorf("failed to load created reward id"))
	}

	return model.Reward{
		Id:         int64(*rewardIdValue),
		FamilyId:   req.FamilyId,
		Name:       req.Name,
		PointsCost: req.PointsCost,
		Stock:      req.Stock,
		Status:     model.RewardStatusActive,
		CreatedBy:  operator.Id,
	}
}

func (x rewardService) ListRewards(userId int64, familyId int64) []model.Reward {
	if MemberService.LoadActiveMember(userId, familyId) == nil {
		panic(fmt.Errorf("permission denied"))
	}

	const sql = `
		SELECT id
			, family_id
			, name
			, points_cost
			, stock
			, status
			, created_by
		FROM rewards
		WHERE family_id=@p1 AND status=@p2
		ORDER BY id DESC
	`
	return resx.Db.Main.MustListOf(model.Reward{}, sql, familyId, model.RewardStatusActive).([]model.Reward)
}
