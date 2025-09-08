package queries

import "go-redis/internal/infrastructure/interface-adapters/sqlite"

type IHMGetMultiUseCase interface {
	Handle(keys map[string][]string) (map[string]map[string]*string, error)
}

type hmgetMultiUseCase struct{ repo sqlite.IRedisRepository }

func NewHMGetMultiUseCase(repo sqlite.IRedisRepository) IHMGetMultiUseCase {
	return &hmgetMultiUseCase{repo: repo}
}

// Optimized multi-key version using single repo call where available
func (h *hmgetMultiUseCase) Handle(keys map[string][]string) (map[string]map[string]*string, error) {
	res := make(map[string]map[string]*string, len(keys))
	flat, err := h.repo.MGetHashFields(keys)
	if err != nil {
		return res, err
	}
	for k, fields := range keys {
		m := make(map[string]*string, len(fields))
		if kv, ok := flat[k]; ok {
			for _, f := range fields {
				if v, ok2 := kv[f]; ok2 {
					vv := v
					m[f] = &vv
				} else {
					m[f] = nil
				}
			}
		} else {
			for _, f := range fields {
				m[f] = nil
			}
		}
		res[k] = m
	}
	return res, nil
}
