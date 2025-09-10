package commands

import "go-redis/internal/infrastructure/interface-adapters/sqlite"

type IHDelUseCase interface { Handle(key, field string) (int, error) }

type hdelUseCase struct{ repo sqlite.IRedisRepository }

func NewHDelUseCase(repo sqlite.IRedisRepository) IHDelUseCase { return &hdelUseCase{repo: repo} }

func (h *hdelUseCase) Handle(key, field string) (int, error) {
	k, err := h.repo.GetKey(key)
	if err != nil { return 0, nil }
	n, err := h.repo.DeleteHashField(k.ID, field)
	if err != nil { return 0, err }
	if n > 0 { _ = h.repo.LogCommand("HDEL", key, []string{field}, "OK") }
	return int(n), nil
}

