package model

import "time"

type PointLogListRequest struct {
	FamilyId int64 `json:"familyId"`
	MemberId int64 `json:"memberId"`
}

type PointLogListItem struct {
	Id          int64           `json:"id"`
	FamilyId    int64           `json:"familyId"`
	MemberId    int64           `json:"memberId"`
	Nickname    string          `json:"nickname"`
	Points      int             `json:"points"`
	SourceType  PointSourceType `json:"sourceType"`
	SourceId    int64           `json:"sourceId"`
	SourceTitle string          `json:"sourceTitle"`
	CreatedAt   time.Time       `json:"createdAt"`
}
