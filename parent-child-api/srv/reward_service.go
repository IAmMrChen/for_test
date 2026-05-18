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

func (x rewardService) ApplyReward(operatorUserId int64, req model.RewardApplyRequest) model.RewardRecord {
	if req.FamilyId == 0 {
		panic(fmt.Errorf("family id is required"))
	}
	if req.RewardId == 0 {
		panic(fmt.Errorf("reward id is required"))
	}

	target := resolveRewardTargetMember(operatorUserId, req.FamilyId, req.MemberId)

	tran := resx.Db.Main.MustCreateTransactionEx()
	defer tran.MustClose()

	reward := loadActiveRewardForUpdate(tran, req.FamilyId, req.RewardId)
	if reward.Stock == 0 {
		panic(fmt.Errorf("reward stock is not enough"))
	}
	if reward.Stock > 0 {
		affected := tran.MustExecute(`
			UPDATE rewards
			SET stock=stock-1
			WHERE id=@p1 AND family_id=@p2 AND status=@p3 AND stock>0
		`, reward.Id, reward.FamilyId, model.RewardStatusActive)
		if affected != 1 {
			panic(fmt.Errorf("reward stock is not enough"))
		}
	}

	affected := tran.MustExecute(`
		UPDATE family_members
		SET current_points=current_points-@p1
		WHERE id=@p2 AND family_id=@p3 AND role_type=@p4 AND status=@p5 AND current_points>=@p1
	`, reward.PointsCost, target.Id, req.FamilyId, model.FamilyRoleChild, model.FamilyMemberStatusActive)
	if affected != 1 {
		panic(fmt.Errorf("points is not enough"))
	}

	tran.MustExecute(`
		INSERT INTO reward_records(family_id, reward_id, member_id, points_cost, status)
		VALUES(@p1, @p2, @p3, @p4, @p5)
	`, req.FamilyId, reward.Id, target.Id, reward.PointsCost, model.RewardRecordStatusApplied)

	recordIdValue, ok := tran.MustScalarInt("SELECT LAST_INSERT_ID()")
	if !ok || recordIdValue == nil {
		panic(fmt.Errorf("failed to load created reward record id"))
	}
	recordId := int64(*recordIdValue)

	tran.MustExecute(`
		INSERT INTO point_logs(family_id, member_id, points, source_type, source_id)
		VALUES(@p1, @p2, @p3, @p4, @p5)
	`, req.FamilyId, target.Id, -reward.PointsCost, model.PointSourceTypeReward, recordId)

	tran.MustCommit()

	return model.RewardRecord{
		Id:         recordId,
		FamilyId:   req.FamilyId,
		RewardId:   reward.Id,
		MemberId:   target.Id,
		PointsCost: reward.PointsCost,
		Status:     model.RewardRecordStatusApplied,
	}
}

func resolveRewardTargetMember(operatorUserId, familyId, memberId int64) model.FamilyMember {
	operator := MemberService.LoadActiveMember(operatorUserId, familyId)
	if operator == nil {
		panic(fmt.Errorf("permission denied"))
	}

	target := *operator
	if memberId != 0 && memberId != operator.Id {
		member := MemberService.LoadActiveMemberById(memberId)
		if member == nil || member.FamilyId != familyId || member.RoleType != model.FamilyRoleChild {
			panic(fmt.Errorf("target child not found"))
		}
		if !memberCanSubmitForChild(*operator, member.IsVirtual) {
			panic(fmt.Errorf("permission denied"))
		}
		target = *member
	}

	if target.RoleType != model.FamilyRoleChild {
		panic(fmt.Errorf("only child can apply reward"))
	}
	return target
}

func loadActiveRewardForUpdate(db structGetter, familyId, rewardId int64) model.Reward {
	const sql = `
		SELECT id
			, family_id
			, name
			, points_cost
			, stock
			, status
			, created_by
		FROM rewards
		WHERE id=@p1 AND family_id=@p2 AND status=@p3
		FOR UPDATE
	`

	reward := &model.Reward{}
	ok := db.MustGetStruct(reward, sql, rewardId, familyId, model.RewardStatusActive)
	if !ok {
		panic(fmt.Errorf("reward not found"))
	}
	return *reward
}
