package interface_adapters

import (
	"go-redis/config"
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
	"log"
)

type ISqliteDbRepositories interface {
	GetRedisRepository() sqlite.IRedisRepository
	Close() error
}

type sqliteDbRepositories struct {
	redisRepository sqlite.IRedisRepository
}

func NewSqliteDbRepositories(config config.IConfig) *sqliteDbRepositories {
	sqliteConn, err := sqlite.NewSQLiteConnection(config)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite: %v", err)
	}
	redisRepo := sqlite.NewRedisRepository(sqliteConn)
	threadSafeRedisRepo := sqlite.NewThreadSafeRedisRepository(redisRepo)
	return &sqliteDbRepositories{
		redisRepository: threadSafeRedisRepo,
	}
}

func (s *sqliteDbRepositories) GetRedisRepository() sqlite.IRedisRepository {
	return s.redisRepository
}

func (s *sqliteDbRepositories) Close() error {
	if c, ok := s.redisRepository.(interface{ Close() error }); ok {
		return c.Close()
	}
	return nil
}
