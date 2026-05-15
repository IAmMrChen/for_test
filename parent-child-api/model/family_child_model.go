package model

type VirtualChildCreateRequest struct {
	FamilyId int64  `json:"familyId"`
	Nickname string `json:"nickname"`
}
