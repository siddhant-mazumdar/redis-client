package commands

import (
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
)

type IMSetUseCase interface {
	Handle(pairs map[string]string) (int, error)
}

type msetUseCase struct{ repo sqlite.IRedisRepository }

func NewMSetUseCase(repo sqlite.IRedisRepository) IMSetUseCase { return &msetUseCase{repo: repo} }

func (m *msetUseCase) Handle(pairs map[string]string) (int, error) {
	if err := m.repo.StoreMultipleStrings(pairs); err != nil {
		return 0, err
	}
	_ = m.repo.LogCommand("MSET", "", []string{}, "OK")
	return len(pairs), nil
}
