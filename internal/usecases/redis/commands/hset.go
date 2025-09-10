package commands

import (
	"fmt"
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
)

type IHSetUseCase interface { Handle(key, field, value string) error }

type hsetUseCase struct{ repo sqlite.IRedisRepository }

func NewHSetUseCase(repo sqlite.IRedisRepository) IHSetUseCase { return &hsetUseCase{repo: repo} }

func (h *hsetUseCase) Handle(key, field, value string) error {
	if key == "" || field == "" {
		return fmt.Errorf("key and field cannot be empty")
	}
	// ensure hash key exists
	_, err := h.repo.StoreKey(key, "hash", 0)
	if err != nil {
		return err
	}
	keyMeta, err := h.repo.GetKey(key)
	if err != nil {
		return err
	}
	if err := h.repo.StoreHash(keyMeta.ID, field, value); err != nil {
		return err
	}
	_ = h.repo.LogCommand("HSET", key, []string{field, value}, "OK")
	return nil
}

