package queries

import (
	"fmt"
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
)

type IGetStoredDataUseCase interface {
	Handle(keyName string) (interface{}, error)
}

type getStoredDataUseCase struct {
	redisRepository sqlite.IRedisRepository
}

func NewGetStoredDataUseCase(redisRepository sqlite.IRedisRepository) IGetStoredDataUseCase {
	return &getStoredDataUseCase{
		redisRepository: redisRepository,
	}
}

func (g *getStoredDataUseCase) Handle(keyName string) (interface{}, error) {
	// Get key metadata
	key, err := g.redisRepository.GetKey(keyName)
	if err != nil {
		return nil, fmt.Errorf("key not found: %w", err)
	}

	// Log the command
	g.redisRepository.LogCommand("GET", keyName, []string{}, "OK")

	// Return data based on key type
	switch key.KeyType {
	case "string":
		value, err := g.redisRepository.GetString(keyName)
		if err != nil {
			return nil, fmt.Errorf("failed to get string value: %w", err)
		}
		return value, nil
	case "hash":
		hashData, err := g.redisRepository.GetHash(keyName)
		if err != nil {
			return nil, fmt.Errorf("failed to get hash value: %w", err)
		}
		return hashData, nil
	default:
		return nil, fmt.Errorf("unsupported key type: %s", key.KeyType)
	}
}
