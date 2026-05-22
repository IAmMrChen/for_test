package model

import "time"

type VirtualChildCreateRequest struct {
	FamilyId int64  `json:"familyId"`
	Nickname string `json:"nickname"`
}

type VirtualChildBindInviteCreateRequest struct {
	FamilyId int64 `json:"familyId"`
	MemberId int64 `json:"memberId"`
}

type VirtualChildBindInviteAcceptRequest struct {
	Token string `json:"token"`
}

type VirtualChildBindInvite struct {
	Id               int64              `json:"id"`
	FamilyId         int64              `json:"familyId"`
	ChildMemberId    int64              `json:"childMemberId"`
	InviterMemberId  int64              `json:"inviterMemberId"`
	Token            string             `json:"token"`
	Status           FamilyInviteStatus `json:"status"`
	ExpiresAt        time.Time          `json:"expiresAt"`
	AcceptedByUserId *int64             `json:"acceptedByUserId"`
	AcceptedAt       *time.Time         `json:"acceptedAt"`
}
