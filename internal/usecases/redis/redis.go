package redis

import (
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
	redis_commands "go-redis/internal/usecases/redis/commands"
	redis_queries "go-redis/internal/usecases/redis/queries"
)

type queries struct {
	GetStoredData     redis_queries.IGetStoredDataUseCase
	GetCommandHistory redis_queries.IGetCommandHistoryUseCase
}

type commands struct {
	ParseRespToKeyValue   redis_commands.IParseRespToKeyValueUseCase
	ParseAndStoreRespData redis_commands.IParseAndStoreRespDataUseCase
	DeleteStoredData      redis_commands.IDeleteStoredDataUseCase
}

type IRedisUseCases interface {
	GetQueries() queries
	GetCommands() commands
}

type redisUseCases struct {
	queries  queries
	commands commands
}

func NewRedisUseCases(redisRepository sqlite.IRedisRepository) IRedisUseCases {
	// Create command use cases
	parseRespToKeyValue := redis_commands.NewParseRespToKeyValueUseCase()
	parseAndStoreRespData := redis_commands.NewParseAndStoreRespDataUseCase(redisRepository, parseRespToKeyValue)
	deleteStoredData := redis_commands.NewDeleteStoredDataUseCase(redisRepository)

	// Create query use cases
	getStoredData := redis_queries.NewGetStoredDataUseCase(redisRepository)
	getCommandHistory := redis_queries.NewGetCommandHistoryUseCase(redisRepository)

	return &redisUseCases{
		queries: queries{
			GetStoredData:     getStoredData,
			GetCommandHistory: getCommandHistory,
		},
		commands: commands{
			ParseRespToKeyValue:   parseRespToKeyValue,
			ParseAndStoreRespData: parseAndStoreRespData,
			DeleteStoredData:      deleteStoredData,
		},
	}
}

func (r *redisUseCases) GetQueries() queries {
	return r.queries
}

func (r *redisUseCases) GetCommands() commands {
	return r.commands
}
