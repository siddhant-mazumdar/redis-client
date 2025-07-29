package interface_adapters

import (
	"go-redis/config"
)

type IInterfaceAdapters interface {
	GetSQLiteDBInterfaceAdapter() ISqliteDbRepositories
}

type interfaceAdapters struct {
	SqliteDbRepositories ISqliteDbRepositories
}

func NewInterfaceAdapters(config config.IConfig) IInterfaceAdapters {
	return &interfaceAdapters{
		SqliteDbRepositories: NewSqliteDbRepositories(config),
	}
}

func (i *interfaceAdapters) GetSQLiteDBInterfaceAdapter() ISqliteDbRepositories {
	return i.SqliteDbRepositories
}
