package main

import "gorm.io/gen"

func exportIndexBackendModels(g *gen.Generator) {
	alarmFilterWord := g.GenerateModelAs("index_backend.alarm_filter_word", "AlarmFilterWord")
	userAuthInfo := g.GenerateModelAs("index_backend.user_auth_info", "UserAuthInfo")
	userLoginLog := g.GenerateModelAs("index_backend.user_login_log", "UserLoginLog")

	g.ApplyBasic(
		alarmFilterWord,
		userAuthInfo,
		userLoginLog,
	)
}
