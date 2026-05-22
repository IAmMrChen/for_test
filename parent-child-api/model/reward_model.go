package model

import "time"

type RewardStatus int

const (
	RewardStatusOffShelf RewardStatus = 0
	RewardStatusActive   RewardStatus = 1
)

type RewardRecordStatus string

const (
	RewardRecordStatusApplied   RewardRecordStatus = "APPLIED"
	RewardRecordStatusDelivered RewardRecordStatus = "DELIVERED"
	RewardRecordStatusReceived  RewardRecordStatus = "RECEIVED"
	RewardRecordStatusRejected  RewardRecordStatus = "REJECTED"
)

func (x RewardRecordStatus) CanOperate() bool {
	return x == RewardRecordStatusApplied
}

func (x RewardRecordStatus) CanReceive() bool {
	return x == RewardRecordStatusDelivered
}

type Reward struct {
	Id         int64        `json:"id"`
	FamilyId   int64        `json:"familyId"`
	Name       string       `json:"name"`
	PointsCost int          `json:"pointsCost"`
	Stock      int          `json:"stock"`
	Status     RewardStatus `json:"status"`
	CreatedBy  int64        `json:"createdBy"`
}

type RewardCreateRequest struct {
	FamilyId   int64  `json:"familyId"`
	Name       string `json:"name"`
	PointsCost int    `json:"pointsCost"`
	Stock      int    `json:"stock"`
}

type RewardUpdateRequest struct {
	FamilyId   int64  `json:"familyId"`
	RewardId   int64  `json:"rewardId"`
	Name       string `json:"name"`
	PointsCost int    `json:"pointsCost"`
	Stock      int    `json:"stock"`
}

type RewardOffShelfRequest struct {
	FamilyId int64 `json:"familyId"`
	RewardId int64 `json:"rewardId"`
}

type RewardApplyRequest struct {
	FamilyId int64 `json:"familyId"`
	RewardId int64 `json:"rewardId"`
	MemberId int64 `json:"memberId"`
}

type RewardRecordOperateRequest struct {
	RecordId int64 `json:"recordId"`
}

type RewardRecordListRequest struct {
	FamilyId int64              `json:"familyId"`
	Status   RewardRecordStatus `json:"status"`
}

type RewardRecord struct {
	Id          int64              `json:"id"`
	FamilyId    int64              `json:"familyId"`
	RewardId    int64              `json:"rewardId"`
	MemberId    int64              `json:"memberId"`
	PointsCost  int                `json:"pointsCost"`
	Status      RewardRecordStatus `json:"status"`
	ApplyTime   time.Time          `json:"applyTime"`
	OperateTime *time.Time         `json:"operateTime"`
	OperateBy   *int64             `json:"operateBy"`
}

type RewardRecordListItem struct {
	Id          int64              `json:"id"`
	FamilyId    int64              `json:"familyId"`
	RewardId    int64              `json:"rewardId"`
	RewardName  string             `json:"rewardName"`
	MemberId    int64              `json:"memberId"`
	Nickname    string             `json:"nickname"`
	PointsCost  int                `json:"pointsCost"`
	Status      RewardRecordStatus `json:"status"`
	ApplyTime   time.Time          `json:"applyTime"`
	OperateTime *time.Time         `json:"operateTime"`
	OperateBy   *int64             `json:"operateBy"`
}
