package queries

import "go-redis/internal/infrastructure/interface-adapters/sqlite"

type IHGetAllUseCase interface { Handle(key string) (map[string]string, error) }

type hgetAllUseCase struct{ repo sqlite.IRedisRepository }

func NewHGetAllUseCase(repo sqlite.IRedisRepository) IHGetAllUseCase { return &hgetAllUseCase{repo: repo} }

func (h *hgetAllUseCase) Handle(key string) (map[string]string, error) {
	return h.repo.GetHash(key)
}

