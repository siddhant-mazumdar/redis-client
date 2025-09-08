package sqlite

import (
	"database/sql"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

type IRedisRepository interface {
	StoreKey(keyName, keyType string, ttl int) (int64, error)
	GetKey(keyName string) (*RedisKey, error)
	DeleteKey(keyName string) error
	StoreString(keyID int64, value string) error
	GetString(keyName string) (string, error)
	MGetStrings(keys []string) (map[string]string, error)
	// Prepared SELECTs caching via stmt getters
	StoreHash(keyID int64, field, value string) error
	StoreHashMap(keyID int64, data map[string]string) error
	StoreMultipleStrings(pairs map[string]string) error
	GetHash(keyName string) (map[string]string, error)
	GetHashFields(keyName string, fields []string) (map[string]string, error)
	MGetHashFields(req map[string][]string) (map[string]map[string]string, error)
	DeleteHashField(keyID int64, field string) (int64, error)
	ListKeys(pattern string) ([]string, error)
	ListKeysPaged(pattern string, offset, count int) ([]string, error)
	DeleteExpiredKeys() (int64, error)
	LogCommand(command, keyName string, args []string, result string) error
	GetCommandHistory(limit int) ([]RedisCommand, error)
	Close() error
}

type RedisRepository struct {
	db *sql.DB
	// Prepared statements for hot paths
	psMu             sync.Mutex
	stmtUpsertKey    *sql.Stmt
	stmtInsertString *sql.Stmt
	stmtInsertHash   *sql.Stmt
	stmtGetKey       *sql.Stmt
	stmtGetString    *sql.Stmt
	stmtGetHashField *sql.Stmt // used by GetHashFields with IN clause; we prepare per-call due to IN size
}

func (r *RedisRepository) Close() error {
	r.psMu.Lock()
	defer r.psMu.Unlock()
	var firstErr error
	closeStmt := func(s **sql.Stmt) {
		if *s != nil {
			if err := (*s).Close(); err != nil && firstErr == nil {
				firstErr = err
			}
			*s = nil
		}
	}
	closeStmt(&r.stmtUpsertKey)
	closeStmt(&r.stmtInsertString)
	closeStmt(&r.stmtInsertHash)
	closeStmt(&r.stmtGetKey)
	closeStmt(&r.stmtGetString)
	return firstErr
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

func (r *RedisRepository) ensureStmtUpsertKey() error {
	r.psMu.Lock()
	defer r.psMu.Unlock()
	if r.stmtUpsertKey == nil {
		stmt, err := r.db.Prepare(`INSERT OR REPLACE INTO redis_keys (key_name, key_type, ttl, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`)
		if err != nil {
			return err
		}
		r.stmtUpsertKey = stmt
	}
	return nil
}

func (r *RedisRepository) ensureStmtInsertString() error {
	r.psMu.Lock()
	defer r.psMu.Unlock()
	if r.stmtInsertString == nil {
		stmt, err := r.db.Prepare(`INSERT OR REPLACE INTO redis_strings (key_id, value) VALUES (?, ?)`)
		if err != nil {
			return err
		}
		r.stmtInsertString = stmt
	}
	return nil
}

func (r *RedisRepository) ensureStmtInsertHash() error {
	r.psMu.Lock()
	defer r.psMu.Unlock()
	if r.stmtInsertHash == nil {
		stmt, err := r.db.Prepare(`INSERT OR REPLACE INTO redis_hashes (key_id, field, value) VALUES (?, ?, ?)`)
		if err != nil {
			return err
		}
		r.stmtInsertHash = stmt
	}
	return nil
}

func (r *RedisRepository) StoreKey(keyName, keyType string, ttl int) (int64, error) {
	if err := r.ensureStmtUpsertKey(); err != nil {
		return 0, err
	}
	result, err := r.stmtUpsertKey.Exec(keyName, keyType, ttl)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func NewRedisRepository(conn *SQLiteConnection) IRedisRepository {
	return &RedisRepository{
		db: conn.GetDB(),
	}
}

func (r *RedisRepository) GetKey(keyName string) (*RedisKey, error) {
	// Cache prepared stmt for key lookup
	r.psMu.Lock()
	stmt := r.stmtGetKey
	r.psMu.Unlock()
	if stmt == nil {
		r.psMu.Lock()
		if r.stmtGetKey == nil {
			var err error
			r.stmtGetKey, err = r.db.Prepare(`SELECT id, key_name, key_type, ttl, created_at, updated_at FROM redis_keys WHERE key_name = ?`)
			if err != nil {
				r.psMu.Unlock()
				return nil, err
			}
		}
		stmt = r.stmtGetKey
		r.psMu.Unlock()
	}
	var key RedisKey
	err := stmt.QueryRow(keyName).Scan(&key.ID, &key.KeyName, &key.KeyType, &key.TTL, &key.CreatedAt, &key.UpdatedAt)
	if err != nil {
		return nil, err
	}
	// TTL enforcement
	if key.TTL >= 0 {
		expireAt := key.UpdatedAt.Add(time.Duration(key.TTL) * time.Second)
		if time.Now().After(expireAt) {
			_ = r.DeleteKey(keyName)
			return nil, sql.ErrNoRows
		}
	}
	return &key, nil
}

func (r *RedisRepository) DeleteKey(keyName string) error {
	query := `DELETE FROM redis_keys WHERE key_name = ?`
	_, err := r.db.Exec(query, keyName)
	return err
}

func (r *RedisRepository) StoreString(keyID int64, value string) error {
	if err := r.ensureStmtInsertString(); err != nil {
		return err
	}
	_, err := r.stmtInsertString.Exec(keyID, value)
	return err
}

func (r *RedisRepository) GetString(keyName string) (string, error) {
	// Cache prepared stmt for string fetch
	r.psMu.Lock()
	stmt := r.stmtGetString
	r.psMu.Unlock()
	if stmt == nil {
		r.psMu.Lock()
		if r.stmtGetString == nil {
			var err error
			r.stmtGetString, err = r.db.Prepare(`SELECT rs.value FROM redis_strings rs JOIN redis_keys rk ON rs.key_id = rk.id WHERE rk.key_name = ?`)
			if err != nil {
				r.psMu.Unlock()
				return "", err
			}
		}
		stmt = r.stmtGetString
		r.psMu.Unlock()
	}
	var value string
	err := stmt.QueryRow(keyName).Scan(&value)
	return value, err
}

func (r *RedisRepository) MGetStrings(keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}
	placeholders := strings.Repeat("?,", len(keys))
	placeholders = placeholders[:len(placeholders)-1]
	query := `
	        SELECT rk.key_name, rs.value
	        FROM redis_strings rs
	        JOIN redis_keys rk ON rs.key_id = rk.id
	        WHERE rk.key_name IN (` + placeholders + `)
	    `
	args := make([]interface{}, len(keys))
	for i, k := range keys {
		args[i] = k
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make(map[string]string, len(keys))
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		res[k] = v
	}
	return res, nil
}

func (r *RedisRepository) StoreHash(keyID int64, field, value string) error {
	if err := r.ensureStmtInsertHash(); err != nil {
		return err
	}
	_, err := r.stmtInsertHash.Exec(keyID, field, value)
	return err
}

func (r *RedisRepository) StoreHashMap(keyID int64, data map[string]string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO redis_hashes (key_id, field, value) VALUES (?, ?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for field, value := range data {
		if _, err := stmt.Exec(keyID, field, value); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// MGetHashFields returns a map: key -> (field -> value), only for requested pairs.
// It executes a single query by constructing dynamic IN clauses for keys and fields,
// and assembles the result in memory.
func (r *RedisRepository) MGetHashFields(req map[string][]string) (map[string]map[string]string, error) {
	if len(req) == 0 {
		return map[string]map[string]string{}, nil
	}
	keys := make([]string, 0, len(req))
	fieldSet := make(map[string]struct{})
	for k, fields := range req {
		keys = append(keys, k)
		for _, f := range fields {
			fieldSet[f] = struct{}{}
		}
	}
	fields := make([]string, 0, len(fieldSet))
	for f := range fieldSet {
		fields = append(fields, f)
	}
	kp := strings.Repeat("?,", len(keys))
	kp = kp[:len(kp)-1]
	fp := strings.Repeat("?,", len(fields))
	fp = fp[:len(fp)-1]
	query := `
	        SELECT rk.key_name, rh.field, rh.value
	        FROM redis_hashes rh
	        JOIN redis_keys rk ON rh.key_id = rk.id
	        WHERE rk.key_name IN (` + kp + `) AND rh.field IN (` + fp + `)
	    `
	args := make([]interface{}, 0, len(keys)+len(fields))
	for _, k := range keys {
		args = append(args, k)
	}
	for _, f := range fields {
		args = append(args, f)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make(map[string]map[string]string, len(keys))
	for rows.Next() {
		var k, f, v string
		if err := rows.Scan(&k, &f, &v); err != nil {
			return nil, err
		}
		m, ok := res[k]
		if !ok {
			m = make(map[string]string)
			res[k] = m
		}
		m[f] = v
	}
	return res, nil
}

func (r *RedisRepository) StoreMultipleStrings(pairs map[string]string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	upsertKey, err := tx.Prepare(`INSERT OR REPLACE INTO redis_keys (key_name, key_type, ttl, updated_at) VALUES (?, 'string', 0, CURRENT_TIMESTAMP)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer upsertKey.Close()
	insertStr, err := tx.Prepare(`INSERT OR REPLACE INTO redis_strings (key_id, value) VALUES (?, ?)`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer insertStr.Close()
	selectID, err := tx.Prepare(`SELECT id FROM redis_keys WHERE key_name = ?`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer selectID.Close()
	for k, v := range pairs {
		if _, err := upsertKey.Exec(k); err != nil {
			tx.Rollback()
			return err
		}
		var id int64
		if err := selectID.QueryRow(k).Scan(&id); err != nil {
			tx.Rollback()
			return err
		}
		if _, err := insertStr.Exec(id, v); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
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

func (r *RedisRepository) GetHashFields(keyName string, fields []string) (map[string]string, error) {
	if len(fields) == 0 {
		return map[string]string{}, nil
	}
	placeholders := strings.Repeat("?,", len(fields))
	placeholders = placeholders[:len(placeholders)-1]
	query := `
	        SELECT rh.field, rh.value
	        FROM redis_hashes rh
	        JOIN redis_keys rk ON rh.key_id = rk.id
	        WHERE rk.key_name = ? AND rh.field IN (` + placeholders + `)
	    `
	args := make([]interface{}, 0, len(fields)+1)
	args = append(args, keyName)
	for _, f := range fields {
		args = append(args, f)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make(map[string]string, len(fields))
	for rows.Next() {
		var f, v string
		if err := rows.Scan(&f, &v); err != nil {
			return nil, err
		}
		res[f] = v
	}
	return res, nil
}

func (r *RedisRepository) DeleteHashField(keyID int64, field string) (int64, error) {
	query := `DELETE FROM redis_hashes WHERE key_id = ? AND field = ?`
	res, err := r.db.Exec(query, keyID, field)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *RedisRepository) ListKeys(pattern string) ([]string, error) {
	like := translatePattern(pattern)
	query := `SELECT key_name FROM redis_keys WHERE key_name LIKE ? ESCAPE '\\'`
	rows, err := r.db.Query(query, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *RedisRepository) ListKeysPaged(pattern string, offset, count int) ([]string, error) {
	if count <= 0 {
		count = 10
	}
	if offset < 0 {
		offset = 0
	}
	like := translatePattern(pattern)
	query := `SELECT key_name FROM redis_keys WHERE key_name LIKE ? ESCAPE '\\' ORDER BY key_name LIMIT ? OFFSET ?`
	rows, err := r.db.Query(query, like, count, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func translatePattern(p string) string {
	// Escape existing LIKE wildcards and backslash
	replacer := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_")
	p = replacer.Replace(p)
	p = strings.ReplaceAll(p, "*", "%")
	p = strings.ReplaceAll(p, "?", "_")
	return p
}

func (r *RedisRepository) DeleteExpiredKeys() (int64, error) {
	query := `DELETE FROM redis_keys WHERE ttl >= 0 AND datetime(updated_at, '+' || ttl || ' seconds') <= CURRENT_TIMESTAMP`
	res, err := r.db.Exec(query)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
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
