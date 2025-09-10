package internal

import (
	"go-redis/config"
	interface_adapters "go-redis/internal/infrastructure/interface-adapters"
	"go-redis/internal/usecases"
)

func NewAdapterUseCaseBridge(config config.IConfig) usecases.IUseCases {
	adapters := interface_adapters.NewInterfaceAdapters(config)

	useCases := usecases.NewUseCases(
		config,
		adapters.GetSQLiteDBInterfaceAdapter().GetRedisRepository(),
	)
	return useCases
}
