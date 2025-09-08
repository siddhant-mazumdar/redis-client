package queries

import "go-redis/internal/infrastructure/interface-adapters/sqlite"

type IMGetUseCase interface {
	Handle(keys []string) ([]interface{}, error)
}

type mgetUseCase struct{ repo sqlite.IRedisRepository }

func NewMGetUseCase(repo sqlite.IRedisRepository) IMGetUseCase { return &mgetUseCase{repo: repo} }

func (m *mgetUseCase) Handle(keys []string) ([]interface{}, error) {
	res := make([]interface{}, len(keys))
	// Single query to fetch all existing string values
	kv, err := m.repo.MGetStrings(keys)
	if err != nil {
		return nil, err
	}
	for i, k := range keys {
		if v, ok := kv[k]; ok {
			res[i] = v
		} else {
			res[i] = nil
		}
	}
	return res, nil
}
