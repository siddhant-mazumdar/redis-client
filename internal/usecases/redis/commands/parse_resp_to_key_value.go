package commands

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type IParseRespToKeyValueUseCase interface {
	Handle(respData string) (map[string]interface{}, error)
}

type parseRespToKeyValueUseCase struct{}

func NewParseRespToKeyValueUseCase() IParseRespToKeyValueUseCase {
	return &parseRespToKeyValueUseCase{}
}

func (p *parseRespToKeyValueUseCase) Handle(respData string) (map[string]interface{}, error) {
	reader := strings.NewReader(respData)
	bufReader := bufio.NewReader(reader)

	result := make(map[string]interface{})

	for {
		value, err := p.parseRespValue(bufReader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Handle key-value pairs from parsed RESP
		if arr, ok := value.([]interface{}); ok && len(arr) >= 2 {
			if key, keyOk := arr[0].(string); keyOk {
				result[key] = arr[1]
			}
		}
	}

	return result, nil
}

func (p *parseRespToKeyValueUseCase) parseRespValue(reader *bufio.Reader) (interface{}, error) {
	typeByte, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	switch typeByte {
	case '+': // Simple String
		return p.parseSimpleString(reader)
	case '-': // Error
		return p.parseError(reader)
	case ':': // Integer
		return p.parseInteger(reader)
	case '$': // Bulk String
		return p.parseBulkString(reader)
	case '*': // Array
		return p.parseArray(reader)
	default:
		return nil, fmt.Errorf("unknown RESP type: %c", typeByte)
	}
}

func (p *parseRespToKeyValueUseCase) parseSimpleString(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(line, "\r\n"), nil
}

func (p *parseRespToKeyValueUseCase) parseError(reader *bufio.Reader) (error, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	return errors.New(strings.TrimSuffix(line, "\r\n")), nil
}

func (p *parseRespToKeyValueUseCase) parseInteger(reader *bufio.Reader) (int64, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSuffix(line, "\r\n"), 10, 64)
}

func (p *parseRespToKeyValueUseCase) parseBulkString(reader *bufio.Reader) (interface{}, error) {
	lengthStr, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	length, err := strconv.Atoi(strings.TrimSuffix(lengthStr, "\r\n"))
	if err != nil {
		return nil, err
	}

	if length == -1 {
		return nil, nil // Null bulk string
	}

	data := make([]byte, length+2) // +2 for \r\n
	_, err = io.ReadFull(reader, data)
	if err != nil {
		return nil, err
	}

	return string(data[:length]), nil
}

func (p *parseRespToKeyValueUseCase) parseArray(reader *bufio.Reader) ([]interface{}, error) {
	lengthStr, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	length, err := strconv.Atoi(strings.TrimSuffix(lengthStr, "\r\n"))
	if err != nil {
		return nil, err
	}

	if length == -1 {
		return nil, nil // Null array
	}

	result := make([]interface{}, length)
	for i := 0; i < length; i++ {
		value, err := p.parseRespValue(reader)
		if err != nil {
			return nil, err
		}
		result[i] = value
	}

	return result, nil
}
