package model

import "time"

type FamilyInviteStatus string

const (
	FamilyInviteStatusActive   FamilyInviteStatus = "ACTIVE"
	FamilyInviteStatusAccepted FamilyInviteStatus = "ACCEPTED"
	FamilyInviteStatusExpired  FamilyInviteStatus = "EXPIRED"
	FamilyInviteStatusCanceled FamilyInviteStatus = "CANCELED"
)

type FamilyInviteCreateRequest struct {
	FamilyId   int64      `json:"familyId"`
	TargetRole FamilyRole `json:"targetRole"`
}

type FamilyInviteAcceptRequest struct {
	Token string `json:"token"`
}

type FamilyInvite struct {
	Id               int64              `json:"id"`
	FamilyId         int64              `json:"familyId"`
	InviterMemberId  int64              `json:"inviterMemberId"`
	TargetRole       FamilyRole         `json:"targetRole"`
	Token            string             `json:"token"`
	Status           FamilyInviteStatus `json:"status"`
	ExpiresAt        time.Time          `json:"expiresAt"`
	AcceptedByUserId *int64             `json:"acceptedByUserId"`
	AcceptedAt       *time.Time         `json:"acceptedAt"`
}
