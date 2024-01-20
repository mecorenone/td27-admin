package base

import (
	"server/service"
)

type ApiGroup struct {
	LogRegApi
	CasbinApi
	JwtApi
}

var (
	jwtService    = service.ServiceGroupApp.Base.JwtService
	logRegService = service.ServiceGroupApp.Base.LogRegService
	casbinService = service.ServiceGroupApp.Base.CasbinService
)
