package queries

import "go-redis/internal/infrastructure/interface-adapters/sqlite"

type IHMGetOptimized interface { Handle(key string, fields []string) ([]interface{}, error) }

type hmgetOptimized struct{ repo sqlite.IRedisRepository }

func NewHMGetOptimized(repo sqlite.IRedisRepository) IHMGetOptimized { return &hmgetOptimized{repo: repo} }

func (h *hmgetOptimized) Handle(key string, fields []string) ([]interface{}, error) {
	res := make([]interface{}, len(fields))
	m, err := h.repo.GetHashFields(key, fields)
	if err != nil { return res, err }
	for i, f := range fields {
		if v, ok := m[f]; ok { res[i] = v } else { res[i] = nil }
	}
	return res, nil
}

