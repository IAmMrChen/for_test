package model

type FamilyRole string

const (
	FamilyRoleOwner  FamilyRole = "OWNER"
	FamilyRoleAdmin  FamilyRole = "ADMIN"
	FamilyRoleParent FamilyRole = "PARENT"
	FamilyRoleChild  FamilyRole = "CHILD"
)

func (x FamilyRole) IsParentRole() bool {
	return x == FamilyRoleOwner || x == FamilyRoleAdmin || x == FamilyRoleParent
}

func (x FamilyRole) CanManageMembers() bool {
	return x == FamilyRoleOwner || x == FamilyRoleAdmin
}

type FamilyMemberStatus string

const (
	FamilyMemberStatusActive  FamilyMemberStatus = "ACTIVE"
	FamilyMemberStatusRemoved FamilyMemberStatus = "REMOVED"
)

type FamilyMember struct {
	Id                int64              `json:"id"`
	FamilyId          int64              `json:"familyId"`
	UserId            *int64             `json:"userId"`
	RoleType          FamilyRole         `json:"roleType"`
	Nickname          string             `json:"nickname"`
	CurrentPoints     int                `json:"currentPoints"`
	TotalEarnedPoints int                `json:"totalEarnedPoints"`
	IsVirtual         bool               `json:"isVirtual"`
	Status            FamilyMemberStatus `json:"status"`
}
