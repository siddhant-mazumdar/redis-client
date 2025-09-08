package tcp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-redis/internal/usecases"
	rtcp "go-redis/internal/usecases/redis/tcp"
)

type Server struct {
	useCases usecases.IUseCases
	ln       net.Listener
	handler  rtcp.ITcpCommandHandler

	mu sync.Mutex
}

func NewServer(useCases usecases.IUseCases) *Server {
	s := &Server{useCases: useCases}
	s.handler = rtcp.NewTcpCommandHandler(useCases.GetRedisUseCases())
	return s
}

// Start launches a Redis-compatible TCP listener on the given port
func (s *Server) Start(port int) {
	addr := ":" + strconv.Itoa(port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("TCP server failed to listen on %s: %v", addr, err)
		return
	}
	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()
	log.Printf("TCP server listening on %s", addr)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				log.Printf("TCP accept error: %v", err)
				continue
			}
			// Best-effort TCP keepalive and per-connection read deadlines
			if tc, ok := conn.(*net.TCPConn); ok {
				_ = tc.SetKeepAlive(true)
				_ = tc.SetKeepAlivePeriod(3 * time.Minute)
			}
			go s.handleConnection(conn)
		}
	}()
}

// Close stops the TCP listener gracefully
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	for {
		// Refresh deadline to avoid idle connections lingering forever
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("TCP read error: %v", err)
			return
		}
		line = strings.TrimSuffix(line, "\r\n")
		if len(line) == 0 {
			continue
		}

		// Prefer RESP array, else support inline protocol
		var args []string
		if strings.HasPrefix(line, "*") {
			argc, err := strconv.Atoi(line[1:])
			if err != nil || argc <= 0 {
				writeError(w, "ERR Protocol error: invalid multibulk length")
				continue
			}

			args = make([]string, 0, argc)
			for i := 0; i < argc; i++ {
				// Expect $<len> then payload
				lenLine, err := r.ReadString('\n')
				if err != nil {
					writeError(w, "ERR Protocol error: invalid bulk length")
					return
				}
				lenLine = strings.TrimSuffix(lenLine, "\r\n")
				if !strings.HasPrefix(lenLine, "$") {
					writeError(w, "ERR Protocol error: expected bulk string")
					return
				}
				blen, err := strconv.Atoi(lenLine[1:])
				if err != nil || blen < 0 {
					writeError(w, "ERR Protocol error: invalid bulk length")
					return
				}
				buf := make([]byte, blen+2)
				if _, err := io.ReadFull(r, buf); err != nil {
					writeError(w, "ERR Protocol error: short bulk read")
					return
				}
				arg := string(buf[:blen])
				args = append(args, arg)
			}
		} else {
			var err error
			args, err = parseInlineArgs(line)
			if err != nil {
				writeError(w, err.Error())
				continue
			}
		}

		if len(args) == 0 {
			writeError(w, "ERR empty command")
			continue
		}

		resp, err := s.handler.Handle(args)
		if err != nil {
			writeError(w, fmt.Sprintf("ERR %v", err))
			continue
		}
		switch resp.Kind {
		case rtcp.SimpleString:
			writeSimpleString(w, resp.Simple)
		case rtcp.BulkString:
			writeBulkString(w, resp.Bulk)
		case rtcp.NullBulk:
			writeNullBulkString(w)
		case rtcp.Integer:
			writeInteger(w, resp.Int)
		case rtcp.Integer64:
			writeInteger64(w, resp.Int64)
		case rtcp.Array:
			writeArray(w, resp.Array)
		case rtcp.Error:
			writeError(w, resp.Err)
		case rtcp.ScanTuple:
			w.WriteString("*2\r\n")
			writeBulkString(w, strconv.Itoa(resp.Cursor))
			arr := make([]interface{}, len(resp.Keys))
			for i, k := range resp.Keys {
				arr[i] = k
			}
			writeArray(w, arr)
		default:
			writeError(w, "ERR unknown command")
		}
	}
}

func writeSimpleString(w *bufio.Writer, s string) {
	w.WriteString("+" + s + "\r\n")
	w.Flush()
}

func writeError(w *bufio.Writer, s string) {
	w.WriteString("-" + s + "\r\n")
	w.Flush()
}

func writeBulkString(w *bufio.Writer, s string) {
	w.WriteString("$" + strconv.Itoa(len(s)) + "\r\n" + s + "\r\n")
	w.Flush()
}

func writeNullBulkString(w *bufio.Writer) {
	w.WriteString("$-1\r\n")
	w.Flush()
}

func writeInteger(w *bufio.Writer, n int) {
	w.WriteString(":" + strconv.Itoa(n) + "\r\n")
	w.Flush()
}

func writeInteger64(w *bufio.Writer, n int64) {
	w.WriteString(":" + strconv.FormatInt(n, 10) + "\r\n")
	w.Flush()
}

func writeArray(w *bufio.Writer, arr []interface{}) {
	w.WriteString("*" + strconv.Itoa(len(arr)) + "\r\n")
	for _, v := range arr {
		switch t := v.(type) {
		case nil:
			writeNullBulkString(w)
		case string:
			writeBulkString(w, t)
		default:
			writeBulkString(w, fmt.Sprintf("%v", t))
		}
	}
	w.Flush()
}

// parseInlineArgs implements a simple Redis inline protocol parser:
// - space-delimited arguments
// - supports double quotes to group with spaces and backslash escapes (\n, \r, \t, \", \\)
func parseInlineArgs(line string) ([]string, error) {
	s := strings.TrimSpace(line)
	args := make([]string, 0, 8)
	var b strings.Builder
	inQuotes := false
	escaped := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped {
			switch c {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '"', '\\':
				b.WriteByte(c)
			default:
				b.WriteByte(c)
			}
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if c == '"' {
			inQuotes = !inQuotes
			continue
		}
		if c == ' ' && !inQuotes {
			if b.Len() > 0 {
				args = append(args, b.String())
				b.Reset()
			}
			continue
		}
		b.WriteByte(c)
	}
	if inQuotes {
		return nil, fmt.Errorf("ERR Protocol error: unmatched quotes")
	}
	if escaped {
		return nil, fmt.Errorf("ERR Protocol error: incomplete escape sequence")
	}
	if b.Len() > 0 {
		args = append(args, b.String())
	}
	return args, nil
}
