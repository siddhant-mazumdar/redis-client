package commands

import (
	"fmt"
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
)

type IExpireUseCase interface { Handle(key string, ttlSeconds int) (int, error) }

type expireUseCase struct{ repo sqlite.IRedisRepository }

func NewExpireUseCase(repo sqlite.IRedisRepository) IExpireUseCase { return &expireUseCase{repo: repo} }

func (e *expireUseCase) Handle(key string, ttlSeconds int) (int, error) {
	keyMeta, err := e.repo.GetKey(key)
	if err != nil {
		return 0, nil // key does not exist per Redis contract
	}
	// StoreKey with same key/type updates updated_at and ttl
	if _, err := e.repo.StoreKey(keyMeta.KeyName, keyMeta.KeyType, ttlSeconds); err != nil {
		return 0, fmt.Errorf("failed to set ttl: %w", err)
	}
	_ = e.repo.LogCommand("EXPIRE", key, []string{fmt.Sprint(ttlSeconds)}, "OK")
	return 1, nil
}

