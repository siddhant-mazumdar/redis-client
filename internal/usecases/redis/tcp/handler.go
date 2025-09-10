package tcp

import (
	"fmt"
	"strconv"
	"strings"

	"go-redis/internal/usecases/redis"
)

type RespKind int

const (
	SimpleString RespKind = iota
	BulkString
	NullBulk
	Integer
	Integer64
	Array
	Error
	ScanTuple
)

type CommandResponse struct {
	Kind   RespKind
	Simple string
	Bulk   string
	Int    int
	Int64  int64
	Array  []interface{}
	Err    string
	// For SCAN
	Cursor int
	Keys   []string
}

type ITcpCommandHandler interface {
	Handle(args []string) (CommandResponse, error)
}

type tcpCommandHandler struct {
	uc redis.IRedisUseCases
}

func NewTcpCommandHandler(uc redis.IRedisUseCases) ITcpCommandHandler {
	return &tcpCommandHandler{uc: uc}
}

func (h *tcpCommandHandler) Handle(args []string) (CommandResponse, error) {
	if len(args) == 0 {
		return CommandResponse{Kind: Error, Err: "ERR empty command"}, nil
	}
	cmd := strings.ToUpper(args[0])
	switch cmd {
	case "PING":
		if len(args) == 1 {
			return CommandResponse{Kind: SimpleString, Simple: "PONG"}, nil
		}
		return CommandResponse{Kind: BulkString, Bulk: args[1]}, nil
	case "ECHO":
		if len(args) != 2 {
			return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'echo'"}, nil
		}
		return CommandResponse{Kind: BulkString, Bulk: args[1]}, nil
	case "SET":
		if len(args) < 3 {
			return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'set'"}, nil
		}
		if err := h.uc.GetCommands().SetString.Handle(args[1], args[2]); err != nil {
			return CommandResponse{Kind: Error, Err: fmt.Sprintf("ERR %v", err)}, nil
		}
		return CommandResponse{Kind: SimpleString, Simple: "OK"}, nil
	case "GET":
		if len(args) != 2 {
			return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'get'"}, nil
		}
		val, err := h.uc.GetQueries().GetStoredData.Handle(args[1])
		if err != nil {
			return CommandResponse{Kind: NullBulk}, nil
		}
		if str, ok := val.(string); ok {
			return CommandResponse{Kind: BulkString, Bulk: str}, nil
		}
		return CommandResponse{Kind: BulkString, Bulk: fmt.Sprintf("%v", val)}, nil
	case "MSET":
		if (len(args)-1)%2 != 0 || len(args) < 3 {
			return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'mset'"}, nil
		}
		pairs := map[string]string{}
		for i := 1; i < len(args); i += 2 {
			pairs[args[i]] = args[i+1]
		}
		if _, err := h.uc.GetCommands().MSet.Handle(pairs); err != nil {
			return CommandResponse{Kind: Error, Err: fmt.Sprintf("ERR %v", err)}, nil
		}
		return CommandResponse{Kind: SimpleString, Simple: "OK"}, nil
	case "MGET":
		if len(args) < 2 {
			return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'mget'"}, nil
		}
		vals, err := h.uc.GetQueries().MGet.Handle(args[1:])
		if err != nil { return CommandResponse{Kind: Error, Err: fmt.Sprintf("ERR %v", err)}, nil }
		return CommandResponse{Kind: Array, Array: vals}, nil
	case "HMSET":
		if (len(args)-2)%2 != 0 || len(args) < 4 {
			return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'hmset'"}, nil
		}
		data := map[string]string{}
		for i := 2; i < len(args); i += 2 { data[args[i]] = args[i+1] }
		if err := h.uc.GetCommands().HMSet.Handle(args[1], data); err != nil {
			return CommandResponse{Kind: Error, Err: fmt.Sprintf("ERR %v", err)}, nil
		}
		return CommandResponse{Kind: SimpleString, Simple: "OK"}, nil
	case "HMGET":
		if len(args) < 3 {
			return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'hmget'"}, nil
		}
		vals, err := h.uc.GetQueries().HMGet.Handle(args[1], args[2:])
		if err != nil { return CommandResponse{Kind: Error, Err: fmt.Sprintf("ERR %v", err)}, nil }
		return CommandResponse{Kind: Array, Array: vals}, nil
	case "DEL":
		if len(args) < 2 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'del'"}, nil }
		deleted := 0
		for _, k := range args[1:] {
			if err := h.uc.GetCommands().DeleteStoredData.Handle(k); err == nil { deleted++ }
		}
		return CommandResponse{Kind: Integer, Int: deleted}, nil
	case "EXISTS":
		if len(args) < 2 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'exists'"}, nil }
		existsCount := 0
		for _, k := range args[1:] {
			ok, _ := h.uc.GetQueries().ExistsKey.Handle(k)
			if ok { existsCount++ }
		}
		return CommandResponse{Kind: Integer, Int: existsCount}, nil
	case "EXPIRE":
		if len(args) != 3 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'expire'"}, nil }
		ttl, err := strconv.Atoi(args[2])
		if err != nil || ttl < 0 { return CommandResponse{Kind: Error, Err: "ERR invalid expire time in seconds"}, nil }
		res, err := h.uc.GetCommands().Expire.Handle(args[1], ttl)
		if err != nil { return CommandResponse{Kind: Integer, Int: 0}, nil }
		return CommandResponse{Kind: Integer, Int: res}, nil
	case "HSET":
		if len(args) != 4 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'hset'"}, nil }
		if err := h.uc.GetCommands().HSet.Handle(args[1], args[2], args[3]); err != nil {
			return CommandResponse{Kind: Integer, Int: 0}, nil
		}
		return CommandResponse{Kind: Integer, Int: 1}, nil
	case "HGET":
		if len(args) != 3 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'hget'"}, nil }
		val, err := h.uc.GetQueries().HGet.Handle(args[1], args[2])
		if err != nil { return CommandResponse{Kind: NullBulk}, nil }
		return CommandResponse{Kind: BulkString, Bulk: val}, nil
	case "HEXISTS":
		if len(args) != 3 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'hexists'"}, nil }
		ok, err := h.uc.GetQueries().HExists.Handle(args[1], args[2])
		if err != nil { return CommandResponse{Kind: Integer, Int: 0}, nil }
		if ok { return CommandResponse{Kind: Integer, Int: 1}, nil }
		return CommandResponse{Kind: Integer, Int: 0}, nil
	case "HGETALL":
		if len(args) != 2 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'hgetall'"}, nil }
		hmap, err := h.uc.GetQueries().HGetAll.Handle(args[1])
		if err != nil { return CommandResponse{Kind: Array, Array: nil}, nil }
		arr := make([]interface{}, 0, len(hmap)*2)
		for f, v := range hmap { arr = append(arr, f, v) }
		return CommandResponse{Kind: Array, Array: arr}, nil
	case "HLEN":
		if len(args) != 2 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'hlen'"}, nil }
		n, err := h.uc.GetQueries().HLEN.Handle(args[1])
		if err != nil { return CommandResponse{Kind: Integer, Int: 0}, nil }
		return CommandResponse{Kind: Integer, Int: n}, nil
	case "HKEYS":
		if len(args) != 2 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'hkeys'"}, nil }
		keys, err := h.uc.GetQueries().HKeys.Handle(args[1])
		if err != nil { return CommandResponse{Kind: Array, Array: nil}, nil }
		arr := make([]interface{}, len(keys))
		for i, k := range keys { arr[i] = k }
		return CommandResponse{Kind: Array, Array: arr}, nil
	case "HVALS":
		if len(args) != 2 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'hvals'"}, nil }
		vals, err := h.uc.GetQueries().HVals.Handle(args[1])
		if err != nil { return CommandResponse{Kind: Array, Array: nil}, nil }
		arr := make([]interface{}, len(vals))
		for i, v := range vals { arr[i] = v }
		return CommandResponse{Kind: Array, Array: arr}, nil
	case "HDEL":
		if len(args) != 3 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'hdel'"}, nil }
		n, err := h.uc.GetCommands().HDel.Handle(args[1], args[2])
		if err != nil { return CommandResponse{Kind: Integer, Int: 0}, nil }
		return CommandResponse{Kind: Integer, Int: n}, nil
	case "INCR":
		if len(args) != 2 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'incr'"}, nil }
		nv, err := h.uc.GetCommands().Incr.Handle(args[1])
		if err != nil { return CommandResponse{Kind: Error, Err: fmt.Sprintf("ERR %v", err)}, nil }
		return CommandResponse{Kind: Integer64, Int64: nv}, nil
	case "KEYS":
		if len(args) != 2 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'keys'"}, nil }
		keys, err := h.uc.GetQueries().ScanKeys.Handle(args[1], 0)
		if err != nil { return CommandResponse{Kind: Error, Err: fmt.Sprintf("ERR %v", err)}, nil }
		arr := make([]interface{}, len(keys))
		for i, k := range keys { arr[i] = k }
		return CommandResponse{Kind: Array, Array: arr}, nil
	case "SCAN":
		if len(args) < 2 { return CommandResponse{Kind: Error, Err: "ERR wrong number of arguments for 'scan'"}, nil }
		cursor, err := strconv.Atoi(args[1])
		if err != nil || cursor < 0 { return CommandResponse{Kind: Error, Err: "ERR invalid cursor"}, nil }
		pattern := "*"
		count := 10
		for i := 2; i < len(args); i++ {
			switch strings.ToUpper(args[i]) {
			case "MATCH":
				if i+1 >= len(args) { return CommandResponse{Kind: Error, Err: "ERR syntax error"}, nil }
				i++
				pattern = args[i]
			case "COUNT":
				if i+1 >= len(args) { return CommandResponse{Kind: Error, Err: "ERR syntax error"}, nil }
				i++
				if v, e := strconv.Atoi(args[i]); e == nil && v > 0 { count = v }
			}
		}
		next, keys, err := h.uc.GetQueries().ScanCursor.Handle(cursor, pattern, count)
		if err != nil { return CommandResponse{Kind: Error, Err: fmt.Sprintf("ERR %v", err)}, nil }
		return CommandResponse{Kind: ScanTuple, Cursor: next, Keys: keys}, nil
	default:
		return CommandResponse{Kind: Error, Err: "ERR unknown command"}, nil
	}
}

