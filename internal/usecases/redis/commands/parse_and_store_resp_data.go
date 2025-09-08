package commands

import (
	"encoding/json"
	"fmt"
	"go-redis/internal/infrastructure/interface-adapters/sqlite"
)

type ParseAndStoreResult struct {
	ParsedData   map[string]interface{} `json:"parsed_data"`
	StoredKeys   []string               `json:"stored_keys"`
	FailedKeys   []string               `json:"failed_keys,omitempty"`
	TotalKeys    int                    `json:"total_keys"`
	SuccessCount int                    `json:"success_count"`
}

type IParseAndStoreRespDataUseCase interface {
	Handle(respData string) (*ParseAndStoreResult, error)
}

type parseAndStoreRespDataUseCase struct {
	redisRepository  sqlite.IRedisRepository
	parseRespUseCase IParseRespToKeyValueUseCase
}

func NewParseAndStoreRespDataUseCase(
	redisRepository sqlite.IRedisRepository,
	parseRespUseCase IParseRespToKeyValueUseCase,
) IParseAndStoreRespDataUseCase {
	return &parseAndStoreRespDataUseCase{
		redisRepository:  redisRepository,
		parseRespUseCase: parseRespUseCase,
	}
}

func (p *parseAndStoreRespDataUseCase) Handle(respData string) (*ParseAndStoreResult, error) {
	// Parse RESP data
	parsedData, err := p.parseRespUseCase.Handle(respData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RESP data: %w", err)
	}

	result := &ParseAndStoreResult{
		ParsedData: parsedData,
		StoredKeys: []string{},
		FailedKeys: []string{},
		TotalKeys:  len(parsedData),
	}

	// Store each key-value pair
	for key, value := range parsedData {
		if err := p.storeKeyValue(key, value); err != nil {
			result.FailedKeys = append(result.FailedKeys, key)
		} else {
			result.StoredKeys = append(result.StoredKeys, key)
			result.SuccessCount++
		}
	}

	// Log the command
	p.redisRepository.LogCommand("PARSE_AND_STORE", "", []string{}, "OK")

	return result, nil
}

func (p *parseAndStoreRespDataUseCase) storeKeyValue(key string, value interface{}) error {
	switch v := value.(type) {
	case string:
		// Store as string
		keyID, err := p.redisRepository.StoreKey(key, "string", 0)
		if err != nil {
			return err
		}
		return p.redisRepository.StoreString(keyID, v)

	case map[string]interface{}:
		// Store as hash
		_, err := p.redisRepository.StoreKey(key, "hash", 0)
		if err != nil {
			return err
		}

		hashData := make(map[string]string)
		for k, val := range v {
			if strVal, ok := val.(string); ok {
				hashData[k] = strVal
			} else {
				// Convert to JSON string if not a string
				jsonBytes, err := json.Marshal(val)
				if err != nil {
					return fmt.Errorf("failed to marshal hash field %s: %w", k, err)
				}
				hashData[k] = string(jsonBytes)
			}
		}
		keyID, err := p.redisRepository.StoreKey(key, "hash", 0)
		if err != nil {
			return err
		}

		if err := p.redisRepository.StoreHashMap(keyID, hashData); err != nil {
			return fmt.Errorf("failed to store hash map: %w", err)
		}
		return nil

	default:
		// Convert other types to JSON string
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}

		keyID, err := p.redisRepository.StoreKey(key, "string", 0)
		if err != nil {
			return err
		}
		return p.redisRepository.StoreString(keyID, string(jsonBytes))
	}
}
