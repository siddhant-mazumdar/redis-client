package queries

import "go-redis/internal/infrastructure/interface-adapters/sqlite"

type IExistsKeyUseCase interface {
	Handle(key string) (bool, error)
}

type existsKeyUseCase struct{ repo sqlite.IRedisRepository }

func NewExistsKeyUseCase(repo sqlite.IRedisRepository) IExistsKeyUseCase {
	return &existsKeyUseCase{repo: repo}
}

func (e *existsKeyUseCase) Handle(key string) (bool, error) {
	if _, err := e.repo.GetKey(key); err != nil {
		return false, nil
	}
	return true, nil
}

