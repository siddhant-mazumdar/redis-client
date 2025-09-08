package commands

import (
	"fmt"
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
	"strconv"
)

type IIncrUseCase interface { Handle(key string) (int64, error) }

type incrUseCase struct{ repo sqlite.IRedisRepository }

func NewIncrUseCase(repo sqlite.IRedisRepository) IIncrUseCase { return &incrUseCase{repo: repo} }

func (i *incrUseCase) Handle(key string) (int64, error) {
	var current int64 = 0
	if k, err := i.repo.GetKey(key); err == nil && k.KeyType == "string" {
		if s, err := i.repo.GetString(key); err == nil {
			if s != "" {
				v, e := strconv.ParseInt(s, 10, 64)
				if e != nil { return 0, fmt.Errorf("value is not an integer or out of range") }
				current = v
			}
		}
	}
	current++
	keyID, err := i.repo.StoreKey(key, "string", 0)
	if err != nil { return 0, err }
	if err := i.repo.StoreString(keyID, strconv.FormatInt(current, 10)); err != nil { return 0, err }
	_ = i.repo.LogCommand("INCR", key, []string{}, "OK")
	return current, nil
}

