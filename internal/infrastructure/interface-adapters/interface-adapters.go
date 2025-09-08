package interface_adapters

import (
	"go-redis/config"
)

type IInterfaceAdapters interface {
	GetSQLiteDBInterfaceAdapter() ISqliteDbRepositories
	Close() error
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

func (i *interfaceAdapters) Close() error {
	if i.SqliteDbRepositories != nil {
		return i.SqliteDbRepositories.Close()
	}
	return nil
}
