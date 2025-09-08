package http

import (
	"context"
	"fmt"
	"go-redis/config"
	"go-redis/internal/infrastructure/input-ports/tcp"
	"go-redis/internal/usecases"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	stdhttp "net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type server struct {
	useCases    usecases.IUseCases
	echo        *echo.Echo
	controllers *Controllers
	config      config.IConfig
	messageBus  *MessageBus
	shutdown    chan os.Signal
	tcpServer   *tcp.Server
}

func NewServer(useCases usecases.IUseCases, config config.IConfig) *server {
	server := &server{
		useCases:    useCases,
		echo:        echo.New(),
		controllers: NewControllers(useCases),
		config:      config,
		messageBus:  NewMessageBus(10), // 10 worker goroutines
		shutdown:    make(chan os.Signal, 1),
	}

	// Configure Echo to handle requests asynchronously
	server.echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: middleware.DefaultSkipper,
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
	}))

	// Remove this line to fix the issue:
	// server.echo.Use(server.asyncRequestMiddleware)

	server.addHealthCheckRoutes()
	server.initializeRoutes()

	// Start message bus
	server.messageBus.Start()

	// Start TCP server for direct connections (Redis-compatible subset)
	// Default port 6379 unless overridden by REDIS_TCP_PORT env var
	server.tcpServer = tcp.NewServer(useCases)
	go func() {
		port := 6379
		if p := os.Getenv("REDIS_TCP_PORT"); p != "" {
			if n, err := strconv.Atoi(p); err == nil {
				port = n
			}
		}
		server.tcpServer.Start(port)
	}()

	return server
}

func (s *server) addHealthCheckRoutes() {
	s.echo.GET("/health-check", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	s.echo.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	s.echo.GET("/", func(c echo.Context) error {
		return c.String(200, "Welcome to redis server")
	})
}

func (s *server) Start() {
	port := 10001
	if os.Getenv("PORT") != "" {
		if portNo, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
			port = portNo
		}
	}

	bindAddr := os.Getenv("BIND_ADDR")
	if bindAddr == "" {
		bindAddr = "127.0.0.1"
	}
	addr := fmt.Sprintf("%s:%d", bindAddr, port)

	// Setup graceful shutdown
	signal.Notify(s.shutdown, os.Interrupt, syscall.SIGTERM)

	// Configure HTTP server with timeouts
	s.echo.HideBanner = true
	s.echo.Server = &stdhttp.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("HTTP server starting on %s", addr)
		if err := s.echo.StartServer(s.echo.Server); err != nil && err != stdhttp.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Run on main thread - wait for shutdown signal
	s.runMainLoop()
}

func (s *server) runMainLoop() {
	log.Println("Server running on main thread, waiting for requests...")

	cleanupTicker := time.NewTicker(30 * time.Minute)
	healthTicker := time.NewTicker(10 * time.Second)
	defer cleanupTicker.Stop()
	defer healthTicker.Stop()

	for {
		select {
		case sig := <-s.shutdown:
			log.Printf("Received signal: %v", sig)
			s.gracefulShutdown()
			return
		case <-healthTicker.C:
			s.performMaintenanceTasks()
		case <-cleanupTicker.C:
			// background TTL cleanup every 30 minutes
			deleted, err := s.useCases.GetRedisUseCases().GetCommands().CleanupExpired.Handle()
			if err != nil {
				log.Printf("TTL cleanup error: %v", err)
			} else if deleted > 0 {
				log.Printf("TTL cleanup removed %d expired keys", deleted)
			}
		}
	}
}

func (s *server) performMaintenanceTasks() {
	// lightweight, do not hold read lock while doing DB work
	// 30-min TTL cleanup cadence handled via ticker in runMainLoop
	log.Printf("Server health check - Time: %v", time.Now())
}

func (s *server) gracefulShutdown() {
	log.Println("Starting graceful shutdown...")

	// Stop message bus
	s.messageBus.Stop()

	// Close TCP server listener
	if s.tcpServer != nil {
		if err := s.tcpServer.Close(); err != nil {
			log.Printf("TCP server close error: %v", err)
		}
	}

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown Echo server
	if err := s.echo.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	// Close use cases (will close repository statements)
	if err := s.useCases.Close(); err != nil {
		log.Printf("Error closing use cases: %v", err)
	}

	log.Println("Server shutdown complete")
}
