package queries

import "go-redis/internal/infrastructure/interface-adapters/sqlite"

type IHexistsUseCase interface { Handle(key, field string) (bool, error) }

type hexistsUseCase struct{ repo sqlite.IRedisRepository }

func NewHexistsUseCase(repo sqlite.IRedisRepository) IHexistsUseCase { return &hexistsUseCase{repo: repo} }

func (h *hexistsUseCase) Handle(key, field string) (bool, error) {
	m, err := h.repo.GetHash(key)
	if err != nil { return false, err }
	_, ok := m[field]
	return ok, nil
}

