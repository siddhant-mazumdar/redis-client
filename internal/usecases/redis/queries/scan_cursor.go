package queries

import "go-redis/internal/infrastructure/interface-adapters/sqlite"

type IScanCursorUseCase interface { Handle(cursor int, pattern string, count int) (nextCursor int, keys []string, err error) }

type scanCursorUseCase struct{ repo sqlite.IRedisRepository }

func NewScanCursorUseCase(repo sqlite.IRedisRepository) IScanCursorUseCase { return &scanCursorUseCase{repo: repo} }

func (s *scanCursorUseCase) Handle(cursor int, pattern string, count int) (int, []string, error) {
	// Simplified: fetch all matching keys, slice per count, return next cursor
	keys, err := s.repo.ListKeys(pattern)
	if err != nil { return 0, nil, err }
	if count <= 0 { count = 10 }
	if cursor < 0 { cursor = 0 }
	end := cursor + count
	if end > len(keys) { end = len(keys) }
	chunk := keys[cursor:end]
	next := 0
	if end < len(keys) { next = end }
	return next, chunk, nil
}

