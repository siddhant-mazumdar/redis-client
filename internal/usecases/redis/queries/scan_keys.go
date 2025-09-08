package queries

import "go-redis/internal/infrastructure/interface-adapters/sqlite"

type IScanKeysUseCase interface { Handle(pattern string, count int) ([]string, error) }

type scanKeysUseCase struct{ repo sqlite.IRedisRepository }

func NewScanKeysUseCase(repo sqlite.IRedisRepository) IScanKeysUseCase { return &scanKeysUseCase{repo: repo} }

func (s *scanKeysUseCase) Handle(pattern string, count int) ([]string, error) {
	// For simplicity ignore count and return all matching keys
	return s.repo.ListKeys(pattern)
}

