package api

import (
	"net/http"

	"parent-child-api/api/apix"
	authapi "parent-child-api/api/auth"
	demoapi "parent-child-api/api/demo"
	familyapi "parent-child-api/api/family"
	inviteapi "parent-child-api/api/invite"
	memberapi "parent-child-api/api/member"
	rewardapi "parent-child-api/api/reward"
	taskapi "parent-child-api/api/task"
)

func RegisterRoutes(mux *http.ServeMux, jwtSecret string) {
	authApi := authapi.AuthApi{JwtSecret: jwtSecret}
	demoApi := demoapi.DemoApi{}
	familyApi := familyapi.FamilyApi{}
	memberApi := memberapi.MemberApi{}
	inviteApi := inviteapi.InviteApi{}
	taskApi := taskapi.TaskApi{}
	rewardApi := rewardapi.RewardApi{}

	mux.HandleFunc("POST /api/auth/demoLogin", authApi.DemoLogin)
	mux.HandleFunc("GET /api/demo/me", apix.WithAuth(jwtSecret, demoApi.Me))
	mux.HandleFunc("POST /api/family/create", apix.WithAuth(jwtSecret, familyApi.Create))
	mux.HandleFunc("GET /api/family/list", apix.WithAuth(jwtSecret, familyApi.List))
	mux.HandleFunc("POST /api/member/createVirtualChild", apix.WithAuth(jwtSecret, memberApi.CreateVirtualChild))
	mux.HandleFunc("POST /api/invite/create", apix.WithAuth(jwtSecret, inviteApi.Create))
	mux.HandleFunc("POST /api/invite/accept", apix.WithAuth(jwtSecret, inviteApi.Accept))
	mux.HandleFunc("POST /api/task/create", apix.WithAuth(jwtSecret, taskApi.Create))
	mux.HandleFunc("GET /api/task/list", apix.WithAuth(jwtSecret, taskApi.List))
	mux.HandleFunc("POST /api/task/claim", apix.WithAuth(jwtSecret, taskApi.Claim))
	mux.HandleFunc("POST /api/task/submit", apix.WithAuth(jwtSecret, taskApi.Submit))
	mux.HandleFunc("POST /api/task/audit", apix.WithAuth(jwtSecret, taskApi.Audit))
	mux.HandleFunc("GET /api/task/records", apix.WithAuth(jwtSecret, taskApi.Records))
	mux.HandleFunc("POST /api/reward/create", apix.WithAuth(jwtSecret, rewardApi.Create))
	mux.HandleFunc("GET /api/reward/list", apix.WithAuth(jwtSecret, rewardApi.List))
	mux.HandleFunc("POST /api/reward/apply", apix.WithAuth(jwtSecret, rewardApi.Apply))
	mux.HandleFunc("POST /api/reward/deliver", apix.WithAuth(jwtSecret, rewardApi.Deliver))
	mux.HandleFunc("POST /api/reward/reject", apix.WithAuth(jwtSecret, rewardApi.Reject))
	mux.HandleFunc("POST /api/reward/receive", apix.WithAuth(jwtSecret, rewardApi.Receive))
	mux.HandleFunc("GET /api/reward/records", apix.WithAuth(jwtSecret, rewardApi.Records))
}
