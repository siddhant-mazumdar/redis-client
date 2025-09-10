package commands

import (
	"fmt"
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
)

type ISetStringUseCase interface {
	Handle(key, value string) error
}

type setStringUseCase struct {
	redisRepository sqlite.IRedisRepository
}

func NewSetStringUseCase(redisRepository sqlite.IRedisRepository) ISetStringUseCase {
	return &setStringUseCase{redisRepository: redisRepository}
}

func (s *setStringUseCase) Handle(key, value string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	keyID, err := s.redisRepository.StoreKey(key, "string", 0)
	if err != nil {
		return err
	}
	if err := s.redisRepository.StoreString(keyID, value); err != nil {
		return err
	}
	_ = s.redisRepository.LogCommand("SET", key, []string{value}, "OK")
	return nil
}

