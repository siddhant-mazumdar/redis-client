package commands

import (
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
)

type ICleanupExpiredKeysUseCase interface {
	Handle() (int64, error)
}

type cleanupExpiredKeysUseCase struct { repo sqlite.IRedisRepository }

func NewCleanupExpiredKeysUseCase(repo sqlite.IRedisRepository) ICleanupExpiredKeysUseCase {
	return &cleanupExpiredKeysUseCase{repo: repo}
}

func (c *cleanupExpiredKeysUseCase) Handle() (int64, error) {
	return c.repo.DeleteExpiredKeys()
}

