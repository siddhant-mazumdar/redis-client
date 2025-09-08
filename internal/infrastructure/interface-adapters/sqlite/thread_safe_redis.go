package sqlite

import (
	"sync"
)

type ThreadSafeRedisRepository struct {
	repo IRedisRepository
	mu   sync.RWMutex
}

func NewThreadSafeRedisRepository(repo IRedisRepository) IRedisRepository {
	return &ThreadSafeRedisRepository{
		repo: repo,
	}
}

func (t *ThreadSafeRedisRepository) StoreKey(keyName, keyType string, ttl int) (int64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.StoreKey(keyName, keyType, ttl)
}

func (t *ThreadSafeRedisRepository) GetKey(keyName string) (*RedisKey, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.repo.GetKey(keyName)
}

func (t *ThreadSafeRedisRepository) DeleteKey(keyName string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.DeleteKey(keyName)
}

func (t *ThreadSafeRedisRepository) StoreString(keyID int64, value string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.StoreString(keyID, value)
}

func (t *ThreadSafeRedisRepository) StoreMultipleStrings(pairs map[string]string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.StoreMultipleStrings(pairs)
}

func (t *ThreadSafeRedisRepository) GetString(keyName string) (string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.repo.GetString(keyName)
}

func (t *ThreadSafeRedisRepository) MGetStrings(keys []string) (map[string]string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.repo.MGetStrings(keys)
}

func (t *ThreadSafeRedisRepository) StoreHash(keyID int64, field, value string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.StoreHash(keyID, field, value)
}

func (t *ThreadSafeRedisRepository) StoreHashMap(keyID int64, data map[string]string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.StoreHashMap(keyID, data)
}

func (t *ThreadSafeRedisRepository) GetHash(keyName string) (map[string]string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.repo.GetHash(keyName)
}

func (t *ThreadSafeRedisRepository) GetHashFields(keyName string, fields []string) (map[string]string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.repo.GetHashFields(keyName, fields)
}

func (t *ThreadSafeRedisRepository) DeleteHashField(keyID int64, field string) (int64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.DeleteHashField(keyID, field)
}

func (t *ThreadSafeRedisRepository) ListKeys(pattern string) ([]string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.repo.ListKeys(pattern)
}

func (t *ThreadSafeRedisRepository) ListKeysPaged(pattern string, offset, count int) ([]string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.repo.ListKeysPaged(pattern, offset, count)
}

func (t *ThreadSafeRedisRepository) DeleteExpiredKeys() (int64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.DeleteExpiredKeys()
}

func (t *ThreadSafeRedisRepository) LogCommand(command, keyName string, args []string, result string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.LogCommand(command, keyName, args, result)
}

func (t *ThreadSafeRedisRepository) GetCommandHistory(limit int) ([]RedisCommand, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.repo.GetCommandHistory(limit)
}

func (t *ThreadSafeRedisRepository) MGetHashFields(req map[string][]string) (map[string]map[string]string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.repo.MGetHashFields(req)
}

func (t *ThreadSafeRedisRepository) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.Close()
}
