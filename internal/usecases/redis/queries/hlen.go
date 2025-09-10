package queries

import "go-redis/internal/infrastructure/interface-adapters/sqlite"

type IHLENUseCase interface { Handle(key string) (int, error) }

type hlenUseCase struct{ repo sqlite.IRedisRepository }

func NewHLENUseCase(repo sqlite.IRedisRepository) IHLENUseCase { return &hlenUseCase{repo: repo} }

func (h *hlenUseCase) Handle(key string) (int, error) {
	m, err := h.repo.GetHash(key)
	if err != nil { return 0, err }
	return len(m), nil
}

