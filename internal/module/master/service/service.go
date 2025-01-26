package service

import (
	"codebase-app/internal/module/master/ports"
)

var _ ports.MasterService = &masterService{}

type masterService struct {
	repo ports.MasterRepository
}

func NewMasterService(repo ports.MasterRepository) *masterService {
	return &masterService{
		repo: repo,
	}
}
