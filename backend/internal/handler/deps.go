package handler

import (
	"gomusical/internal/auth"
	"gomusical/internal/config"
	"gomusical/internal/download"
	"gomusical/internal/repo"
	"gomusical/internal/service"
	"gomusical/internal/storage"
	"gomusical/internal/transcode"
)

type API struct {
	Cfg     config.Config
	Repos   *repo.Repos
	Sess    *auth.Store
	Store   *storage.Local
	Access  *service.Access
	Sponsor *service.Sponsor
	Worker  *transcode.Worker
	Limiter *download.Limiter
	Eng     *transcode.Engine
}
