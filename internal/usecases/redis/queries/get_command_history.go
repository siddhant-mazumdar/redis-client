package queries

import (
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
)

type IGetCommandHistoryUseCase interface {
	Handle(limit int) ([]sqlite.RedisCommand, error)
}

type getCommandHistoryUseCase struct {
	redisRepository sqlite.IRedisRepository
}

func NewGetCommandHistoryUseCase(redisRepository sqlite.IRedisRepository) IGetCommandHistoryUseCase {
	return &getCommandHistoryUseCase{
		redisRepository: redisRepository,
	}
}

func (g *getCommandHistoryUseCase) Handle(limit int) ([]sqlite.RedisCommand, error) {
	return g.redisRepository.GetCommandHistory(limit)
}
