package queries

import "go-redis/internal/infrastructure/interface-adapters/sqlite"

type IHMGetUseCase interface {
	Handle(key string, fields []string) ([]interface{}, error)
}

type hmgetUseCase struct{ repo sqlite.IRedisRepository }

func NewHMGetUseCase(repo sqlite.IRedisRepository) IHMGetUseCase { return &hmgetUseCase{repo: repo} }

func (h *hmgetUseCase) Handle(key string, fields []string) ([]interface{}, error) {
	res := make([]interface{}, len(fields))
	m, err := h.repo.GetHashFields(key, fields)
	if err != nil {
		return res, err
	}
	for i, f := range fields {
		if v, ok := m[f]; ok {
			res[i] = v
		} else {
			res[i] = nil
		}
	}
	return res, nil
}
