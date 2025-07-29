package http

import (
	"go-redis/internal/infrastructure/input-ports/http/redis"
	"go-redis/internal/usecases"
)

type Controllers struct {
	RespController redis.IRespController
}

func NewControllers(useCases usecases.IUseCases) *Controllers {
	redisUseCases := useCases.GetRedisUseCases()
	return &Controllers{
		RespController: redis.NewRespController(redisUseCases),
	}
}
