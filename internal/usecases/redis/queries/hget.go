package queries

import (
	"fmt"
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
)

type IHGetUseCase interface { Handle(key, field string) (string, error) }

type hgetUseCase struct{ repo sqlite.IRedisRepository }

func NewHGetUseCase(repo sqlite.IRedisRepository) IHGetUseCase { return &hgetUseCase{repo: repo} }

func (h *hgetUseCase) Handle(key, field string) (string, error) {
	keyMeta, err := h.repo.GetKey(key)
	if err != nil {
		return "", fmt.Errorf("key not found")
	}
	if keyMeta.KeyType != "hash" {
		return "", fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	m, err := h.repo.GetHash(key)
	if err != nil {
		return "", err
	}
	v, ok := m[field]
	if !ok {
		return "", fmt.Errorf("field not found")
	}
	return v, nil
}

