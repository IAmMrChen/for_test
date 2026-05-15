package model

type Family struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	CreatorId int64  `json:"creatorId"`
}

type FamilyCreateRequest struct {
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
}

type FamilyCreateResponse struct {
	Family Family       `json:"family"`
	Member FamilyMember `json:"member"`
}

type FamilyListItem struct {
	FamilyId   int64      `json:"familyId"`
	FamilyName string     `json:"familyName"`
	MemberId   int64      `json:"memberId"`
	RoleType   FamilyRole `json:"roleType"`
	Nickname   string     `json:"nickname"`
}
