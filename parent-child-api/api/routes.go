package api

import (
	"net/http"

	"parent-child-api/api/apix"
	authapi "parent-child-api/api/auth"
	dashboardapi "parent-child-api/api/dashboard"
	demoapi "parent-child-api/api/demo"
	familyapi "parent-child-api/api/family"
	inviteapi "parent-child-api/api/invite"
	memberapi "parent-child-api/api/member"
	rewardapi "parent-child-api/api/reward"
	taskapi "parent-child-api/api/task"
)

func RegisterRoutes(mux *http.ServeMux, jwtSecret string) {
	authApi := authapi.AuthApi{JwtSecret: jwtSecret}
	dashboardApi := dashboardapi.DashboardApi{}
	demoApi := demoapi.DemoApi{}
	familyApi := familyapi.FamilyApi{}
	memberApi := memberapi.MemberApi{}
	inviteApi := inviteapi.InviteApi{}
	taskApi := taskapi.TaskApi{}
	rewardApi := rewardapi.RewardApi{}

	mux.HandleFunc("POST /api/auth/demoLogin", recoverRoute(authApi.DemoLogin))
	mux.HandleFunc("GET /api/dashboard/summary", recoverRoute(apix.WithAuth(jwtSecret, dashboardApi.Summary)))
	mux.HandleFunc("GET /api/demo/me", recoverRoute(apix.WithAuth(jwtSecret, demoApi.Me)))
	mux.HandleFunc("POST /api/family/create", recoverRoute(apix.WithAuth(jwtSecret, familyApi.Create)))
	mux.HandleFunc("GET /api/family/list", recoverRoute(apix.WithAuth(jwtSecret, familyApi.List)))
	mux.HandleFunc("POST /api/member/createVirtualChild", recoverRoute(apix.WithAuth(jwtSecret, memberApi.CreateVirtualChild)))
	mux.HandleFunc("GET /api/member/list", recoverRoute(apix.WithAuth(jwtSecret, memberApi.List)))
	mux.HandleFunc("POST /api/invite/create", recoverRoute(apix.WithAuth(jwtSecret, inviteApi.Create)))
	mux.HandleFunc("POST /api/invite/accept", recoverRoute(apix.WithAuth(jwtSecret, inviteApi.Accept)))
	mux.HandleFunc("POST /api/task/create", recoverRoute(apix.WithAuth(jwtSecret, taskApi.Create)))
	mux.HandleFunc("GET /api/task/list", recoverRoute(apix.WithAuth(jwtSecret, taskApi.List)))
	mux.HandleFunc("POST /api/task/claim", recoverRoute(apix.WithAuth(jwtSecret, taskApi.Claim)))
	mux.HandleFunc("POST /api/task/submit", recoverRoute(apix.WithAuth(jwtSecret, taskApi.Submit)))
	mux.HandleFunc("POST /api/task/audit", recoverRoute(apix.WithAuth(jwtSecret, taskApi.Audit)))
	mux.HandleFunc("GET /api/task/records", recoverRoute(apix.WithAuth(jwtSecret, taskApi.Records)))
	mux.HandleFunc("POST /api/reward/create", recoverRoute(apix.WithAuth(jwtSecret, rewardApi.Create)))
	mux.HandleFunc("GET /api/reward/list", recoverRoute(apix.WithAuth(jwtSecret, rewardApi.List)))
	mux.HandleFunc("POST /api/reward/apply", recoverRoute(apix.WithAuth(jwtSecret, rewardApi.Apply)))
	mux.HandleFunc("POST /api/reward/deliver", recoverRoute(apix.WithAuth(jwtSecret, rewardApi.Deliver)))
	mux.HandleFunc("POST /api/reward/reject", recoverRoute(apix.WithAuth(jwtSecret, rewardApi.Reject)))
	mux.HandleFunc("POST /api/reward/receive", recoverRoute(apix.WithAuth(jwtSecret, rewardApi.Receive)))
	mux.HandleFunc("GET /api/reward/records", recoverRoute(apix.WithAuth(jwtSecret, rewardApi.Records)))
}

func recoverRoute(handler http.HandlerFunc) http.HandlerFunc {
	return apix.WithRecover(handler)
}
