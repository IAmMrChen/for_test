package api

import (
	"net/http"

	"parent-child-api/api/apix"
	authapi "parent-child-api/api/auth"
	demoapi "parent-child-api/api/demo"
	familyapi "parent-child-api/api/family"
	inviteapi "parent-child-api/api/invite"
	memberapi "parent-child-api/api/member"
)

func RegisterRoutes(mux *http.ServeMux, jwtSecret string) {
	authApi := authapi.AuthApi{JwtSecret: jwtSecret}
	demoApi := demoapi.DemoApi{}
	familyApi := familyapi.FamilyApi{}
	memberApi := memberapi.MemberApi{}
	inviteApi := inviteapi.InviteApi{}

	mux.HandleFunc("POST /api/auth/demoLogin", authApi.DemoLogin)
	mux.HandleFunc("GET /api/demo/me", apix.WithAuth(jwtSecret, demoApi.Me))
	mux.HandleFunc("POST /api/family/create", apix.WithAuth(jwtSecret, familyApi.Create))
	mux.HandleFunc("GET /api/family/list", apix.WithAuth(jwtSecret, familyApi.List))
	mux.HandleFunc("POST /api/member/createVirtualChild", apix.WithAuth(jwtSecret, memberApi.CreateVirtualChild))
	mux.HandleFunc("POST /api/invite/create", apix.WithAuth(jwtSecret, inviteApi.Create))
	mux.HandleFunc("POST /api/invite/accept", apix.WithAuth(jwtSecret, inviteApi.Accept))
}
