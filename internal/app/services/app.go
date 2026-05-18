package services

import "context"

type AppService struct{}

func NewAppService() *AppService {
	return &AppService{}
}

func (s *AppService) Health(ctx context.Context) string {
	return "ok"
}
