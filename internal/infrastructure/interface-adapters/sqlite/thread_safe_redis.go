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

func (t *ThreadSafeRedisRepository) GetString(keyName string) (string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.repo.GetString(keyName)
}

func (t *ThreadSafeRedisRepository) StoreHash(keyID int64, field, value string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.repo.StoreHash(keyID, field, value)
}

func (t *ThreadSafeRedisRepository) GetHash(keyName string) (map[string]string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.repo.GetHash(keyName)
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
