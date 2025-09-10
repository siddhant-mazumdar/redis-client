package commands

import "go-redis/internal/infrastructure/interface-adapters/sqlite"

type IHMSetUseCase interface { Handle(key string, data map[string]string) error }

type hmsetUseCase struct{ repo sqlite.IRedisRepository }

func NewHMSetUseCase(repo sqlite.IRedisRepository) IHMSetUseCase { return &hmsetUseCase{repo: repo} }

func (h *hmsetUseCase) Handle(key string, data map[string]string) error {
	// Ensure key exists and is hash
	_, err := h.repo.StoreKey(key, "hash", 0)
	if err != nil { return err }
	k, err := h.repo.GetKey(key)
	if err != nil { return err }
	return h.repo.StoreHashMap(k.ID, data)
}

