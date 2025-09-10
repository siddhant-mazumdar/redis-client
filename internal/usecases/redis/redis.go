package redis

import (
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
	redis_commands "go-redis/internal/usecases/redis/commands"
	redis_queries "go-redis/internal/usecases/redis/queries"
)

type queries struct {
	GetStoredData     redis_queries.IGetStoredDataUseCase
	GetCommandHistory redis_queries.IGetCommandHistoryUseCase
	ExistsKey         redis_queries.IExistsKeyUseCase
	HGet              redis_queries.IHGetUseCase
	MGet              redis_queries.IMGetUseCase
	HMGet             redis_queries.IHMGetUseCase
	ScanKeys          redis_queries.IScanKeysUseCase
	HExists           redis_queries.IHexistsUseCase
	HGetAll           redis_queries.IHGetAllUseCase
	HLEN              redis_queries.IHLENUseCase
	HKeys             redis_queries.IHKeysUseCase
	HVals             redis_queries.IHValsUseCase
	HMGetMulti        redis_queries.IHMGetMultiUseCase

	ScanCursor redis_queries.IScanCursorUseCase
}

type commands struct {
	ParseRespToKeyValue   redis_commands.IParseRespToKeyValueUseCase
	ParseAndStoreRespData redis_commands.IParseAndStoreRespDataUseCase
	DeleteStoredData      redis_commands.IDeleteStoredDataUseCase
	SetString             redis_commands.ISetStringUseCase
	Expire                redis_commands.IExpireUseCase
	HSet                  redis_commands.IHSetUseCase
	HDel                  redis_commands.IHDelUseCase
	HMSet                 redis_commands.IHMSetUseCase
	MSet                  redis_commands.IMSetUseCase
	Incr                  redis_commands.IIncrUseCase
	CleanupExpired        redis_commands.ICleanupExpiredKeysUseCase
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
	setString := redis_commands.NewSetStringUseCase(redisRepository)
	expire := redis_commands.NewExpireUseCase(redisRepository)
	hset := redis_commands.NewHSetUseCase(redisRepository)
	hdel := redis_commands.NewHDelUseCase(redisRepository)
	hmset := redis_commands.NewHMSetUseCase(redisRepository)
	mset := redis_commands.NewMSetUseCase(redisRepository)
	incr := redis_commands.NewIncrUseCase(redisRepository)
	cleanup := redis_commands.NewCleanupExpiredKeysUseCase(redisRepository)

	// Create query use cases
	getStoredData := redis_queries.NewGetStoredDataUseCase(redisRepository)
	getCommandHistory := redis_queries.NewGetCommandHistoryUseCase(redisRepository)
	existsKey := redis_queries.NewExistsKeyUseCase(redisRepository)
	hget := redis_queries.NewHGetUseCase(redisRepository)
	mget := redis_queries.NewMGetUseCase(redisRepository)
	hmget := redis_queries.NewHMGetUseCase(redisRepository)
	hmgetMulti := redis_queries.NewHMGetMultiUseCase(redisRepository)
	scan := redis_queries.NewScanKeysUseCase(redisRepository)
	hexists := redis_queries.NewHexistsUseCase(redisRepository)
	hgetall := redis_queries.NewHGetAllUseCase(redisRepository)
	hlen := redis_queries.NewHLENUseCase(redisRepository)
	hkeys := redis_queries.NewHKeysUseCase(redisRepository)
	hvals := redis_queries.NewHValsUseCase(redisRepository)
	scanCursor := redis_queries.NewScanCursorUseCase(redisRepository)

	return &redisUseCases{
		queries: queries{
			GetStoredData:     getStoredData,
			GetCommandHistory: getCommandHistory,
			ExistsKey:         existsKey,
			HGet:              hget,
			MGet:              mget,
			HMGet:             hmget,
			ScanKeys:          scan,
			HExists:           hexists,
			HGetAll:           hgetall,
			HLEN:              hlen,
			HKeys:             hkeys,
			HVals:             hvals,
			ScanCursor:        scanCursor,
			HMGetMulti:        hmgetMulti,
		},
		commands: commands{
			ParseRespToKeyValue:   parseRespToKeyValue,
			ParseAndStoreRespData: parseAndStoreRespData,
			DeleteStoredData:      deleteStoredData,
			SetString:             setString,
			Expire:                expire,
			HSet:                  hset,
			HDel:                  hdel,
			HMSet:                 hmset,
			MSet:                  mset,
			Incr:                  incr,
			CleanupExpired:        cleanup,
		},
	}
}

func (r *redisUseCases) GetQueries() queries {
	return r.queries
}

func (r *redisUseCases) GetCommands() commands {
	return r.commands
}
