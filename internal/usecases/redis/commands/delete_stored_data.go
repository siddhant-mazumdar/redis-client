package commands

import (
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
)

type IDeleteStoredDataUseCase interface {
	Handle(keyName string) error
}

type deleteStoredDataUseCase struct {
	redisRepository sqlite.IRedisRepository
}

func NewDeleteStoredDataUseCase(redisRepository sqlite.IRedisRepository) IDeleteStoredDataUseCase {
	return &deleteStoredDataUseCase{
		redisRepository: redisRepository,
	}
}

func (d *deleteStoredDataUseCase) Handle(keyName string) error {
	// Log the command
	d.redisRepository.LogCommand("DELETE", keyName, []string{}, "OK")

	return d.redisRepository.DeleteKey(keyName)
}
