package sqlite

import (
	"database/sql"
	"encoding/json"
	"time"
)

type IRedisRepository interface {
	StoreKey(keyName, keyType string, ttl int) (int64, error)
	GetKey(keyName string) (*RedisKey, error)
	DeleteKey(keyName string) error
	StoreString(keyID int64, value string) error
	GetString(keyName string) (string, error)
	StoreHash(keyID int64, field, value string) error
	GetHash(keyName string) (map[string]string, error)
	LogCommand(command, keyName string, args []string, result string) error
	GetCommandHistory(limit int) ([]RedisCommand, error)
}

type RedisRepository struct {
	db *sql.DB
}

type RedisKey struct {
	ID        int64     `json:"id"`
	KeyName   string    `json:"key_name"`
	KeyType   string    `json:"key_type"`
	TTL       int       `json:"ttl"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RedisCommand struct {
	ID         int64     `json:"id"`
	Command    string    `json:"command"`
	KeyName    string    `json:"key_name"`
	Arguments  []string  `json:"arguments"`
	Result     string    `json:"result"`
	ExecutedAt time.Time `json:"executed_at"`
}

func NewRedisRepository(conn *SQLiteConnection) IRedisRepository {
	return &RedisRepository{
		db: conn.GetDB(),
	}
}

func (r *RedisRepository) StoreKey(keyName, keyType string, ttl int) (int64, error) {
	query := `
        INSERT OR REPLACE INTO redis_keys (key_name, key_type, ttl, updated_at)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP)
    `
	result, err := r.db.Exec(query, keyName, keyType, ttl)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *RedisRepository) GetKey(keyName string) (*RedisKey, error) {
	query := `SELECT id, key_name, key_type, ttl, created_at, updated_at FROM redis_keys WHERE key_name = ?`

	var key RedisKey
	err := r.db.QueryRow(query, keyName).Scan(
		&key.ID, &key.KeyName, &key.KeyType, &key.TTL, &key.CreatedAt, &key.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *RedisRepository) DeleteKey(keyName string) error {
	query := `DELETE FROM redis_keys WHERE key_name = ?`
	_, err := r.db.Exec(query, keyName)
	return err
}

func (r *RedisRepository) StoreString(keyID int64, value string) error {
	query := `INSERT OR REPLACE INTO redis_strings (key_id, value) VALUES (?, ?)`
	_, err := r.db.Exec(query, keyID, value)
	return err
}

func (r *RedisRepository) GetString(keyName string) (string, error) {
	query := `
        SELECT rs.value 
        FROM redis_strings rs 
        JOIN redis_keys rk ON rs.key_id = rk.id 
        WHERE rk.key_name = ?
    `
	var value string
	err := r.db.QueryRow(query, keyName).Scan(&value)
	return value, err
}

func (r *RedisRepository) StoreHash(keyID int64, field, value string) error {
	query := `INSERT OR REPLACE INTO redis_hashes (key_id, field, value) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, keyID, field, value)
	return err
}

func (r *RedisRepository) GetHash(keyName string) (map[string]string, error) {
	query := `
        SELECT rh.field, rh.value 
        FROM redis_hashes rh 
        JOIN redis_keys rk ON rh.key_id = rk.id 
        WHERE rk.key_name = ?
    `
	rows, err := r.db.Query(query, keyName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var field, value string
		if err := rows.Scan(&field, &value); err != nil {
			return nil, err
		}
		result[field] = value
	}
	return result, nil
}

func (r *RedisRepository) LogCommand(command, keyName string, args []string, result string) error {
	argsJSON, _ := json.Marshal(args)
	query := `
        INSERT INTO redis_commands (command, key_name, arguments, result)
        VALUES (?, ?, ?, ?)
    `
	_, err := r.db.Exec(query, command, keyName, string(argsJSON), result)
	return err
}

func (r *RedisRepository) GetCommandHistory(limit int) ([]RedisCommand, error) {
	query := `
        SELECT id, command, key_name, arguments, result, executed_at
        FROM redis_commands
        ORDER BY executed_at DESC
        LIMIT ?
    `
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []RedisCommand
	for rows.Next() {
		var cmd RedisCommand
		var argsJSON string
		err := rows.Scan(&cmd.ID, &cmd.Command, &cmd.KeyName, &argsJSON, &cmd.Result, &cmd.ExecutedAt)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(argsJSON), &cmd.Arguments)
		commands = append(commands, cmd)
	}
	return commands, nil
}
