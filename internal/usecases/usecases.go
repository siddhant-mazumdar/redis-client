package usecases

import (
	"go-redis/config"
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
	redis_usecases "go-redis/internal/usecases/redis"
)

type IUseCases interface {
	GetRedisUseCases() redis_usecases.IRedisUseCases
}
type UseCases struct {
	RedisUseCases redis_usecases.IRedisUseCases
}

func NewUseCases(
	config config.IConfig,
	redisRepository sqlite.IRedisRepository,
) UseCases {

	return UseCases{
		RedisUseCases: redis_usecases.NewRedisUseCases(redisRepository),
	}
}

func (u UseCases) GetRedisUseCases() redis_usecases.IRedisUseCases {
	return u.RedisUseCases
}
