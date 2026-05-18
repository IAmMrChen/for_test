package model

type DashboardSummary struct {
	FamilyId                 int64      `json:"familyId"`
	FamilyName               string     `json:"familyName"`
	MemberId                 int64      `json:"memberId"`
	RoleType                 FamilyRole `json:"roleType"`
	Nickname                 string     `json:"nickname"`
	CurrentPoints            int        `json:"currentPoints"`
	TotalEarnedPoints        int        `json:"totalEarnedPoints"`
	ActiveTaskCount          int        `json:"activeTaskCount"`
	MyClaimedTaskCount       int        `json:"myClaimedTaskCount"`
	MyPendingTaskCount       int        `json:"myPendingTaskCount"`
	FamilyPendingTaskCount   int        `json:"familyPendingTaskCount"`
	ActiveRewardCount        int        `json:"activeRewardCount"`
	MyAppliedRewardCount     int        `json:"myAppliedRewardCount"`
	FamilyAppliedRewardCount int        `json:"familyAppliedRewardCount"`
}
