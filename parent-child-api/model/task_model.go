package model

import "time"

type TaskCycleType string

const (
	TaskCycleTypeOnce   TaskCycleType = "ONCE"
	TaskCycleTypeDaily  TaskCycleType = "DAILY"
	TaskCycleTypeWeekly TaskCycleType = "WEEKLY"
)

func (x TaskCycleType) Valid() bool {
	return x == TaskCycleTypeOnce || x == TaskCycleTypeDaily || x == TaskCycleTypeWeekly
}

type TaskStatus int

const (
	TaskStatusArchived TaskStatus = 0
	TaskStatusActive   TaskStatus = 1
)

type TaskRecordStatus string

const (
	TaskRecordStatusClaimed  TaskRecordStatus = "CLAIMED"
	TaskRecordStatusPending  TaskRecordStatus = "PENDING"
	TaskRecordStatusApproved TaskRecordStatus = "APPROVED"
	TaskRecordStatusRejected TaskRecordStatus = "REJECTED"
)

type TaskClaimStatus string

const (
	TaskClaimStatusActive  TaskClaimStatus = "ACTIVE"
	TaskClaimStatusStopped TaskClaimStatus = "STOPPED"
)

func (x TaskRecordStatus) CanAudit() bool {
	return x == TaskRecordStatusPending
}

func (x TaskRecordStatus) CanSubmit() bool {
	return x == TaskRecordStatusClaimed
}

type PointSourceType string

const (
	PointSourceTypeTask   PointSourceType = "TASK"
	PointSourceTypeReward PointSourceType = "REWARD"
	PointSourceTypeAdjust PointSourceType = "ADJUST"
)

type Task struct {
	Id        int64         `json:"id"`
	FamilyId  int64         `json:"familyId"`
	Title     string        `json:"title"`
	Points    int           `json:"points"`
	CycleType TaskCycleType `json:"cycleType"`
	Status    TaskStatus    `json:"status"`
	CreatedBy int64         `json:"createdBy"`
}

type TaskCreateRequest struct {
	FamilyId  int64         `json:"familyId"`
	Title     string        `json:"title"`
	Points    int           `json:"points"`
	CycleType TaskCycleType `json:"cycleType"`
}

type TaskUpdateRequest struct {
	FamilyId  int64         `json:"familyId"`
	TaskId    int64         `json:"taskId"`
	Title     string        `json:"title"`
	Points    int           `json:"points"`
	CycleType TaskCycleType `json:"cycleType"`
}

type TaskArchiveRequest struct {
	FamilyId int64 `json:"familyId"`
	TaskId   int64 `json:"taskId"`
}

type TaskClaimListRequest struct {
	FamilyId int64 `json:"familyId"`
}

type TaskListRequest struct {
	FamilyId int64 `json:"familyId"`
}

type TaskClaimRequest struct {
	FamilyId int64 `json:"familyId"`
	TaskId   int64 `json:"taskId"`
	MemberId int64 `json:"memberId"`
}

type TaskSubmitRequest struct {
	FamilyId     int64  `json:"familyId"`
	RecordId     int64  `json:"recordId"`
	TaskId       int64  `json:"taskId"`
	MemberId     int64  `json:"memberId"`
	SubmitRemark string `json:"submitRemark"`
}

type TaskAuditRequest struct {
	RecordId    int64  `json:"recordId"`
	Approved    bool   `json:"approved"`
	AuditRemark string `json:"auditRemark"`
}

type TaskRecordListRequest struct {
	FamilyId int64            `json:"familyId"`
	Status   TaskRecordStatus `json:"status"`
}

type TaskRecord struct {
	Id           int64            `json:"id"`
	FamilyId     int64            `json:"familyId"`
	TaskId       int64            `json:"taskId"`
	MemberId     int64            `json:"memberId"`
	Status       TaskRecordStatus `json:"status"`
	SubmitRemark string           `json:"submitRemark"`
	SubmitTime   time.Time        `json:"submitTime"`
	AuditTime    *time.Time       `json:"auditTime"`
	AuditBy      *int64           `json:"auditBy"`
	AuditRemark  string           `json:"auditRemark"`
}

type TaskClaim struct {
	Id        int64           `json:"id"`
	FamilyId  int64           `json:"familyId"`
	TaskId    int64           `json:"taskId"`
	MemberId  int64           `json:"memberId"`
	Status    TaskClaimStatus `json:"status"`
	ClaimedAt time.Time       `json:"claimedAt"`
	StoppedAt *time.Time      `json:"stoppedAt"`
}

type TaskClaimListItem struct {
	Id       int64           `json:"id"`
	FamilyId int64           `json:"familyId"`
	TaskId   int64           `json:"taskId"`
	MemberId int64           `json:"memberId"`
	Status   TaskClaimStatus `json:"status"`
}

type TaskRecordListItem struct {
	Id           int64            `json:"id"`
	FamilyId     int64            `json:"familyId"`
	TaskId       int64            `json:"taskId"`
	TaskTitle    string           `json:"taskTitle"`
	Points       int              `json:"points"`
	MemberId     int64            `json:"memberId"`
	Nickname     string           `json:"nickname"`
	Status       TaskRecordStatus `json:"status"`
	SubmitRemark string           `json:"submitRemark"`
	SubmitTime   time.Time        `json:"submitTime"`
	AuditTime    *time.Time       `json:"auditTime"`
	AuditBy      *int64           `json:"auditBy"`
	AuditRemark  string           `json:"auditRemark"`
}
