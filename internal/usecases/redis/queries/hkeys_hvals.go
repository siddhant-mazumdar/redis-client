package queries

import "go-redis/internal/infrastructure/interface-adapters/sqlite"

type IHKeysUseCase interface { Handle(key string) ([]string, error) }

type IHValsUseCase interface { Handle(key string) ([]string, error) }

type hkeysUseCase struct{ repo sqlite.IRedisRepository }

type hvalsUseCase struct{ repo sqlite.IRedisRepository }

func NewHKeysUseCase(repo sqlite.IRedisRepository) IHKeysUseCase { return &hkeysUseCase{repo: repo} }

func NewHValsUseCase(repo sqlite.IRedisRepository) IHValsUseCase { return &hvalsUseCase{repo: repo} }

func (h *hkeysUseCase) Handle(key string) ([]string, error) {
	m, err := h.repo.GetHash(key)
	if err != nil { return nil, err }
	keys := make([]string, 0, len(m))
	for k := range m { keys = append(keys, k) }
	return keys, nil
}

func (h *hvalsUseCase) Handle(key string) ([]string, error) {
	m, err := h.repo.GetHash(key)
	if err != nil { return nil, err }
	vals := make([]string, 0, len(m))
	for _, v := range m { vals = append(vals, v) }
	return vals, nil
}

